package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/karthikbalasubramani/netpilot-device-management/internal/auth"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/config"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/database"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/repository"
)

const (
	databaseSetupTimeout  = 10 * time.Second
	adminBootstrapTimeout = 15 * time.Second
)

const (
	adminNameEnv     = "NETPILOT_BOOTSTRAP_ADMIN_NAME"
	adminEmailEnv    = "NETPILOT_BOOTSTRAP_ADMIN_EMAIL"
	adminPasswordEnv = "NETPILOT_BOOTSTRAP_ADMIN_PASSWORD"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"admin bootstrap failed: %v\n",
			err,
		)

		os.Exit(1)
	}
}

func run() error {
	adminName := strings.TrimSpace(
		os.Getenv(adminNameEnv),
	)

	adminEmail := strings.TrimSpace(
		os.Getenv(adminEmailEnv),
	)

	adminPassword := os.Getenv(
		adminPasswordEnv,
	)

	if adminName == "" {
		return fmt.Errorf(
			"%s is required",
			adminNameEnv,
		)
	}

	if adminEmail == "" {
		return fmt.Errorf(
			"%s is required",
			adminEmailEnv,
		)
	}

	if adminPassword == "" {
		return fmt.Errorf(
			"%s is required",
			adminPasswordEnv,
		)
	}

	cfg := config.Load()

	if err := cfg.ValidateEnvConfiguration(); err != nil {
		return fmt.Errorf(
			"validate application configuration: %w",
			err,
		)
	}

	authConfig, err :=
		config.LoadAuthConfig()
	if err != nil {
		return fmt.Errorf(
			"load authentication configuration: %w",
			err,
		)
	}

	mongoDB, err :=
		database.ConnectMongoDB(
			cfg,
		)
	if err != nil {
		return fmt.Errorf(
			"connect to MongoDB: %w",
			err,
		)
	}

	defer func() {
		if err := database.Disconnect(
			mongoDB,
		); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"warning: failed to disconnect from MongoDB: %v\n",
				err,
			)
		}
	}()

	userSetupContext, cancelUserSetup :=
		context.WithTimeout(
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
		return fmt.Errorf(
			"initialize users collection: %w",
			err,
		)
	}

	userCollection :=
		database.UserCollection(
			mongoDB.Client,
			cfg.MongoDatabase,
		)

	userRepository, err :=
		repository.NewUserRepository(
			userCollection,
		)
	if err != nil {
		return fmt.Errorf(
			"initialize user repository: %w",
			err,
		)
	}

	passwordHasher, err :=
		auth.NewPasswordHasher(
			authConfig.BcryptCost,
		)
	if err != nil {
		return fmt.Errorf(
			"initialize password hasher: %w",
			err,
		)
	}

	adminBootstrapper :=
		auth.NewAdminBootstrapper(
			userRepository,
			passwordHasher,
		)

	bootstrapContext, cancelBootstrap :=
		context.WithTimeout(
			context.Background(),
			adminBootstrapTimeout,
		)

	defer cancelBootstrap()

	adminUser, err :=
		adminBootstrapper.Bootstrap(
			bootstrapContext,
			auth.AdminBootstrapRequest{
				Name:     adminName,
				Email:    adminEmail,
				Password: adminPassword,
			},
		)
	if err != nil {
		if errors.Is(
			err,
			auth.ErrAdminAlreadyBootstrapped,
		) {
			return auth.ErrAdminAlreadyBootstrapped
		}

		return fmt.Errorf(
			"bootstrap administrator: %w",
			err,
		)
	}

	fmt.Println(
		"NetPilot administrator created successfully",
	)

	fmt.Printf(
		"user_id: %s\n",
		adminUser.UserID,
	)

	fmt.Printf(
		"email: %s\n",
		adminUser.Email,
	)

	fmt.Printf(
		"role: %s\n",
		adminUser.Role,
	)

	return nil
}
