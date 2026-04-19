package jwtmanager

import (
    "crypto/rsa"
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

type TokenInfo struct {
    UserID string
    Role   string
    Valid  bool
}

type userClaims struct {
    jwt.RegisteredClaims
    UserID string `json:"user_id"`
    Role   string `json:"role"`
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
    duration time.Duration,
) (string, error) {
    if m.privateKey == nil {
        return "", fmt.Errorf("private key is not set, cannot generate token")
    }

    claims := userClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
        UserID: userID,
        Role:   role,
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
) (TokenInfo, error) {
    if m.publicKey == nil {
        return TokenInfo{Valid: false}, fmt.Errorf("public key is not set, cannot verify token")
    }

    t, err := jwt.ParseWithClaims(
        tokenStr,
        &userClaims{},
        func(t *jwt.Token) (interface{}, error) {
            if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
            }
            return m.publicKey, nil
        },
    )

    if err != nil {
        return TokenInfo{Valid: false}, fmt.Errorf("token verification failed: %w", err)
    }

    claims, ok := t.Claims.(*userClaims)
    if !ok || !t.Valid {
        return TokenInfo{Valid: false}, fmt.Errorf("invalid token claims")
    }

    return TokenInfo{
        UserID: claims.UserID,
        Role:   claims.Role,
        Valid:  true,
    }, nil
}