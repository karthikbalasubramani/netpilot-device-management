package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/karthikbalasubramani/netpilot-device-management/internal/auth"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/config"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/database"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/logger"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/repository"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/server"
)

const (
	httpServerShutdownTimeout = 10 * time.Second
	databaseSetupTimeout      = 10 * time.Second
)

// Run initializes application dependencies and starts the NetPilot API server.
func Run() error {
	// Load .env values and application configuration.
	cfg := config.Load()

	// Initialize the terminal application logger before validating the
	// configuration so validation failures can be logged.
	logger.Init(cfg.LogLevel)

	if err := cfg.ValidateEnvConfiguration(); err != nil {
		logger.Error(
			"Application configuration validation failed",
			"error", err,
		)

		return fmt.Errorf(
			"validate application configuration: %w",
			err,
		)
	}

	logger.Debug(
		"Configuration loaded from environment variables successfully",
	)

	// Initialize the dedicated API access file logger.
	apiLogConfig, err := config.LoadAPILogConfig()
	if err != nil {
		logger.Error(
			"Failed to load API log configuration",
			"error", err,
		)

		return fmt.Errorf(
			"load API log configuration: %w",
			err,
		)
	}

	if err := logger.InitAPILogger(logger.APILoggerConfig{
		Enabled:    apiLogConfig.Enabled,
		FilePath:   apiLogConfig.FilePath,
		MaxSizeMB:  apiLogConfig.MaxSizeMB,
		MaxBackups: apiLogConfig.MaxBackups,
		MaxAgeDays: apiLogConfig.MaxAgeDays,
		Compress:   apiLogConfig.Compress,
	}); err != nil {
		logger.Error(
			"Failed to initialize API file logger",
			"error", err,
		)

		return fmt.Errorf(
			"initialize API file logger: %w",
			err,
		)
	}

	defer func() {
		if err := logger.CloseAPILogger(); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"failed to close API file logger: %v\n",
				err,
			)
		}
	}()

	// Load and validate HTTP server configuration after godotenv has loaded
	// values from .env.
	httpServerConfig, err := config.LoadHTTPServerConfig()
	if err != nil {
		logger.Error(
			"Failed to load HTTP server configuration",
			"error", err,
		)

		return fmt.Errorf(
			"load HTTP server configuration: %w",
			err,
		)
	}

	logger.Debug(
		"HTTP server configuration loaded",
		"read_header_timeout",
		httpServerConfig.ReadHeaderTimeout.String(),
		"read_timeout",
		httpServerConfig.ReadTimeout.String(),
		"write_timeout",
		httpServerConfig.WriteTimeout.String(),
		"idle_timeout",
		httpServerConfig.IdleTimeout.String(),
		"max_header_bytes",
		httpServerConfig.MaxHeaderBytes,
		"max_request_body_bytes",
		httpServerConfig.MaxRequestBodyBytes,
	)

	// Load health-probe configuration.
	healthProbeConfig, err := config.LoadHealthProbeConfig()
	if err != nil {
		logger.Error("Failed to load health probe configuration", "error", err)
		return fmt.Errorf("load health probe configuration: %w", err)
	}
	logger.Debug(
		"Health probe configuration loaded",
		"readiness_timeout",
		healthProbeConfig.ReadinessTimeout.String(),
	)

	authConfig, err := config.LoadAuthConfig()
	if err != nil {
		logger.Error(
			"Failed to load authentication configuration",
			"error", err,
		)

		return fmt.Errorf(
			"load authentication configuration: %w",
			err,
		)
	}

	logger.Debug(
		"Authentication configuration loaded",
		"bcrypt_cost",
		authConfig.BcryptCost,
		"jwt_issuer",
		authConfig.JWTIssuer,
		"jwt_audience",
		authConfig.JWTAudience,
		"access_token_ttl",
		authConfig.JWTAccessTokenTTLMinutes.String(),
	)

	logger.Info(
		"Starting NetPilot API",
		"application_name", cfg.AppName,
		"environment", cfg.AppEnv,
		"port", cfg.AppPort,
	)

	// Establish the MongoDB connection during application startup.
	mongoDB, err := database.ConnectMongoDB(cfg)
	if err != nil {
		logger.Error(
			"Failed to connect to MongoDB",
			"error", err,
		)

		return fmt.Errorf(
			"connect to MongoDB: %w",
			err,
		)
	}

	// Register MongoDB disconnection before starting the HTTP server.
	defer func() {
		if err := database.Disconnect(mongoDB); err != nil {
			logger.Error(
				"Failed to disconnect from MongoDB",
				"error", err,
			)

			return
		}

		logger.Info("MongoDB disconnected successfully")
	}()

	logger.Info(
		"MongoDB connected successfully",
		"database", cfg.MongoDatabase,
	)

	// Prepare prerequisits MongoDB collections
	databaseSetupContext, cancelDatabaseSetup := context.WithTimeout(
		context.Background(),
		databaseSetupTimeout,
	)

	err = database.EnsureDeviceCollection(
		databaseSetupContext,
		mongoDB.Client,
		cfg.MongoDatabase,
	)

	cancelDatabaseSetup()

	if err != nil {
		logger.Error(
			"Initialize devices collection failed: %w", err,
		)
		return fmt.Errorf(
			"Initialize devices collection failed: %w",
			err,
		)
	}

	logger.Info(
		"Device collection initialized successfully",
		"database", cfg.MongoDatabase,
		"collection", database.DeviceCollectionName,
	)

	// Prepare the User collection and authentication-related indexes.
	userSetupContext, cancelUserSetup := context.WithTimeout(
		context.Background(),
		databaseSetupTimeout,
	)

	err = database.EnsureUserCollection(
		userSetupContext,
		mongoDB.Client,
		cfg.MongoDatabase,
	)

	cancelUserSetup()

	if err != nil {
		logger.Error(
			"Failed to initialize users collection",
			"error", err,
		)

		return fmt.Errorf(
			"initialize users collection: %w",
			err,
		)
	}

	logger.Info(
		"User collection initialized successfully",
		"database", cfg.MongoDatabase,
		"collection", database.UserCollectionName,
	)

	userCollection := database.UserCollection(
		mongoDB.Client,
		cfg.MongoDatabase,
	)

	userRepository, err := repository.NewUserRepository(
		userCollection,
	)
	if err != nil {
		logger.Error(
			"Failed to initialize user repository",
			"error", err,
		)

		return fmt.Errorf(
			"initialize user repository: %w",
			err,
		)
	}

	passwordHasher, err := auth.NewPasswordHasher(
		authConfig.BcryptCost,
	)
	if err != nil {
		logger.Error(
			"Failed to initialize password hasher",
			"error", err,
		)

		return fmt.Errorf(
			"initialize password hasher: %w",
			err,
		)
	}

	accessTokenIssuer, err := auth.NewAccessTokenIssuer(
		authConfig.JWTSecret,
		authConfig.JWTIssuer,
		authConfig.JWTAudience,
		authConfig.JWTAccessTokenTTLMinutes,
	)
	if err != nil {
		logger.Error(
			"Failed to initialize JWT access token issuer",
			"error", err,
		)

		return fmt.Errorf(
			"initialize JWT access token issuer: %w",
			err,
		)
	}

	authService, err := auth.NewService(
		userRepository,
		passwordHasher,
		accessTokenIssuer,
	)
	if err != nil {
		logger.Error(
			"Failed to initialize authentication service",
			"error", err,
		)

		return fmt.Errorf(
			"initialize authentication service: %w",
			err,
		)
	}

	logger.Debug(
		"Authentication dependencies initialized successfully",
	)

	readinessCheck := func(ctx context.Context) error {
		return database.CheckMongoDBReadiness(ctx, mongoDB.Client)
	}

	httpServer, err := server.NewHTTPServer(
		cfg, httpServerConfig, healthProbeConfig, readinessCheck, authService,
	)

	if err != nil {
		logger.Error(
			"Failed to initialize HTTP server",
			"error", err,
		)

		return fmt.Errorf(
			"initialize HTTP server: %w",
			err,
		)
	}

	serverErrorChannel := make(chan error, 1)

	go func() {
		if err := httpServer.StartHTTPServer(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			serverErrorChannel <- err
		}
	}()

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(
		interruptChannel,
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer signal.Stop(interruptChannel)

	select {
	case err := <-serverErrorChannel:
		logger.Error(
			"HTTP server stopped unexpectedly",
			"error", err,
		)

		return fmt.Errorf(
			"HTTP server failed: %w",
			err,
		)

	case receivedSignal := <-interruptChannel:
		logger.Info(
			"Shutdown signal received",
			"signal", receivedSignal.String(),
		)

		shutdownContext, cancel := context.WithTimeout(
			context.Background(),
			httpServerShutdownTimeout,
		)

		shutdownErr := httpServer.ShutdownHTTPServer(
			shutdownContext,
		)
		cancel()

		if shutdownErr != nil {
			logger.Error(
				"Failed to shut down HTTP server gracefully",
				"error", shutdownErr,
			)

			return fmt.Errorf(
				"shut down HTTP server gracefully: %w",
				shutdownErr,
			)
		}

		logger.Info("HTTP server shut down successfully")

		return nil
	}
}
