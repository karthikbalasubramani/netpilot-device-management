package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/karthikbalasubramani/netpilot-device-management/internal/user"
)

const accessTokenType = "Bearer"

var ErrInvalidAccessToken = errors.New(
	"invalid access token",
)

// AccessToken contains a signed JWT and its expiration information.
type AccessToken struct {
	Token     string
	ExpiresAt time.Time
}

// AccessTokenIssuer defines JWT generation required by authentication.
type AccessTokenIssuer interface {
	Generate(
		userID string,
		role user.Role,
	) (*AccessToken, error)
}

// AccessTokenVerifier defines JWT validation required by protected APIs.
type AccessTokenVerifier interface {
	Verify(
		tokenString string,
	) (*Claims, error)
}

// AccessTokenManager combines access-token generation and validation.
type AccessTokenManager interface {
	AccessTokenIssuer
	AccessTokenVerifier
}

// Claims contains NetPilot-specific JWT claims together with standard
// registered JWT claims.
type Claims struct {
	Role user.Role `json:"role"`

	jwt.RegisteredClaims
}

// jwtAccessTokenManager generates and validates HS256 JWT access tokens.
type jwtAccessTokenManager struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
}

// NewAccessTokenManager creates the JWT access-token manager.
func NewAccessTokenManager(
	secret string,
	issuer string,
	audience string,
	ttl time.Duration,
) (AccessTokenManager, error) {
	secret = strings.TrimSpace(secret)
	issuer = strings.TrimSpace(issuer)
	audience = strings.TrimSpace(audience)

	if secret == "" {
		return nil, fmt.Errorf(
			"JWT signing secret is required",
		)
	}

	if issuer == "" {
		return nil, fmt.Errorf(
			"JWT issuer is required",
		)
	}

	if audience == "" {
		return nil, fmt.Errorf(
			"JWT audience is required",
		)
	}

	if ttl <= 0 {
		return nil, fmt.Errorf(
			"JWT access token TTL must be greater than zero",
		)
	}

	return &jwtAccessTokenManager{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
		ttl:      ttl,
	}, nil
}

// Generate creates a signed JWT access token for an authenticated user.
func (manager *jwtAccessTokenManager) Generate(
	userID string,
	role user.Role,
) (*AccessToken, error) {
	userID = strings.TrimSpace(userID)

	if userID == "" {
		return nil, fmt.Errorf(
			"user ID is required for access token",
		)
	}

	if !isValidRole(role) {
		return nil, fmt.Errorf(
			"valid user role is required for access token",
		)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(manager.ttl)

	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: manager.issuer,

			Subject: userID,

			Audience: jwt.ClaimStrings{
				manager.audience,
			},

			ExpiresAt: jwt.NewNumericDate(
				expiresAt,
			),

			NotBefore: jwt.NewNumericDate(
				now,
			),

			IssuedAt: jwt.NewNumericDate(
				now,
			),

			ID: uuid.NewString(),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	signedToken, err := token.SignedString(
		manager.secret,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"sign JWT access token: %w",
			err,
		)
	}

	return &AccessToken{
		Token:     signedToken,
		ExpiresAt: expiresAt,
	}, nil
}

// Verify parses and validates an incoming JWT access token.
func (manager *jwtAccessTokenManager) Verify(
	tokenString string,
) (*Claims, error) {
	tokenString = strings.TrimSpace(tokenString)

	if tokenString == "" {
		return nil, ErrInvalidAccessToken
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return manager.secret, nil
		},
		jwt.WithValidMethods(
			[]string{
				jwt.SigningMethodHS256.Alg(),
			},
		),
		jwt.WithIssuer(
			manager.issuer,
		),
		jwt.WithAudience(
			manager.audience,
		),
		jwt.WithExpirationRequired(),
		jwt.WithNotBeforeRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: %v",
			ErrInvalidAccessToken,
			err,
		)
	}

	if token == nil || !token.Valid {
		return nil, ErrInvalidAccessToken
	}

	if err := validateAccessTokenClaims(
		claims,
	); err != nil {
		return nil, fmt.Errorf(
			"%w: %v",
			ErrInvalidAccessToken,
			err,
		)
	}

	return claims, nil
}

func validateAccessTokenClaims(
	claims *Claims,
) error {
	if claims == nil {
		return fmt.Errorf(
			"claims are required",
		)
	}

	if strings.TrimSpace(claims.Subject) == "" {
		return fmt.Errorf(
			"token subject is required",
		)
	}

	if strings.TrimSpace(claims.ID) == "" {
		return fmt.Errorf(
			"token ID is required",
		)
	}

	if claims.IssuedAt == nil {
		return fmt.Errorf(
			"token issued-at time is required",
		)
	}

	if !isValidRole(claims.Role) {
		return fmt.Errorf(
			"token contains invalid user role",
		)
	}

	return nil
}

func isValidRole(
	role user.Role,
) bool {
	switch role {
	case user.RoleAdmin,
		user.RoleOperator,
		user.RoleViewer:
		return true

	default:
		return false
	}
}
