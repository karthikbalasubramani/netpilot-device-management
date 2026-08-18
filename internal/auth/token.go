package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/user"
)

const accessTokenType = "Bearer"

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

// Claims contains NetPilot-specific JWT claims together with standard JWT
// registered claims.
type Claims struct {
	Role user.Role `json:"role"`

	jwt.RegisteredClaims
}

// jwtAccessTokenIssuer creates HS256-signed JWT access tokens.
type jwtAccessTokenIssuer struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
}

// NewAccessTokenIssuer creates a JWT access-token issuer.
func NewAccessTokenIssuer(
	secret string,
	issuer string,
	audience string,
	ttl time.Duration,
) (AccessTokenIssuer, error) {

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
	return &jwtAccessTokenIssuer{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
		ttl:      ttl,
	}, nil
}

// Generate creates a signed JWT access token for an authenticated user.
func (issuer *jwtAccessTokenIssuer) Generate(
	userID string,
	role user.Role,
) (*AccessToken, error) {
	userID = strings.TrimSpace(userID)

	if userID == "" {
		return nil, fmt.Errorf(
			"user ID is required for access token",
		)
	}

	if role == "" {
		return nil, fmt.Errorf(
			"user role is required for access token",
		)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(issuer.ttl)

	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: issuer.issuer,

			Subject: userID,

			Audience: jwt.ClaimStrings{
				issuer.audience,
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
		issuer.secret,
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
