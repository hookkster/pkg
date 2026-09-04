package jwtmanager

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	TokenAccess  TokenType = "access"
	TokenRefresh TokenType = "refresh"
)

type TokenInfo struct {
	UserID string
	Role   string
}

type userClaims struct {
	jwt.RegisteredClaims
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	TokenType TokenType `json:"typ"`
}

type JWTManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewManager(
	privatePEM,
	publicPEM []byte,
) (*JWTManager, error) {
	m := &JWTManager{}
	var err error

	if len(privatePEM) > 0 {
		m.privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	if len(publicPEM) > 0 {
		m.publicKey, err = jwt.ParseRSAPublicKeyFromPEM(publicPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
	}

	return m, nil
}

func (m *JWTManager) Generate(
	userID string,
	role string,
	tokenType TokenType,
	duration time.Duration,
) (string, error) {
	if m.privateKey == nil {
		return "", ErrNoPrivateKey
	}

	claims := userClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:    userID,
		Role:      role,
		TokenType: tokenType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(m.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (m *JWTManager) Verify(
	tokenStr string,
	expected TokenType,
) (TokenInfo, error) {
	if m.publicKey == nil {
		return TokenInfo{}, ErrNoPublicKey
	}

	t, err := jwt.ParseWithClaims(
		tokenStr,
		&userClaims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return m.publicKey, nil
		},
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return TokenInfo{}, ErrTokenExpired
		}

		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return TokenInfo{}, ErrInvalidSignature
		}

		return TokenInfo{}, fmt.Errorf("token verification failed: %w", err)
	}

	claims, ok := t.Claims.(*userClaims)
	if !ok || !t.Valid {
		return TokenInfo{}, ErrInvalidToken
	}

	if expected != claims.TokenType {
		return TokenInfo{}, ErrWrongTokenType
	}

	return TokenInfo{
		UserID: claims.UserID,
		Role:   claims.Role,
	}, nil
}
