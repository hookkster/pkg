package jwtmanager_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"testing"
	"time"

	"github.com/hookkster/pkg/jwtmanager"
)

func generateKeyPair(t *testing.T) (privatePEM, publicPEM []byte) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	privateDER := x509.MarshalPKCS1PrivateKey(key)
	privatePEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateDER,
	})

	publicDER := x509.MarshalPKCS1PublicKey(&key.PublicKey)
	publicPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicDER,
	})

	return privatePEM, publicPEM
}

func newTestManager(t *testing.T) *jwtmanager.JWTManager {
	t.Helper()

	privatePEM, publicPEM := generateKeyPair(t)

	m, err := jwtmanager.NewManager(privatePEM, publicPEM)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	return m
}

func TestVerifyRejectsWrongTokenType(t *testing.T) {
	m := newTestManager(t)

	refreshToken, err := m.Generate("user-1", "user", jwtmanager.TokenRefresh, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = m.Verify(refreshToken, jwtmanager.TokenAccess)

	if !errors.Is(err, jwtmanager.ErrWrongTokenType) {
		t.Errorf("err = %v, want ErrWrongTokenType", err)
	}
}

func TestVerifyExpiredToken(t *testing.T) {
	m := newTestManager(t)

	accessToken, err := m.Generate("user-1", "user", jwtmanager.TokenAccess, -time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = m.Verify(accessToken, jwtmanager.TokenAccess)

	if !errors.Is(err, jwtmanager.ErrTokenExpired) {
		t.Errorf("err = %v, want ErrTokenExpired", err)
	}
}

func TestVerifyUserInfoFromToken(t *testing.T) {
	m := newTestManager(t)

	accessToken, err := m.Generate("user-1", "user", jwtmanager.TokenAccess, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tokenInfo, err := m.Verify(accessToken, jwtmanager.TokenAccess)

	if tokenInfo.Role != "user" {
		t.Errorf("toke role = %v, want user", tokenInfo.Role)
	}

	if tokenInfo.UserID != "user-1" {
		t.Errorf("toke role = %v, want user-1", tokenInfo.UserID)
	}
}

func TestVerifyDiffManager(t *testing.T) {
	m1 := newTestManager(t)
	m2 := newTestManager(t)

	accessTokenM1, err := m1.Generate("user-1", "user", jwtmanager.TokenAccess, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	accessTokenM2, err := m2.Generate("user-1", "user", jwtmanager.TokenAccess, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = m1.Verify(accessTokenM2, jwtmanager.TokenAccess)
	if !errors.Is(err, jwtmanager.ErrInvalidSignature) {
		t.Errorf("err = %v, want ErrInvalidToken", err)
	}

	_, err = m2.Verify(accessTokenM1, jwtmanager.TokenAccess)
	if !errors.Is(err, jwtmanager.ErrInvalidSignature) {
		t.Errorf("err = %v, want ErrInvalidToken", err)
	}
}

func TestGenerateWithoutPrivateKey(t *testing.T) {
	_, publicPEM := generateKeyPair(t)

	m, err := jwtmanager.NewManager(nil, publicPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = m.Generate("user-1", "user", jwtmanager.TokenAccess, time.Hour)

	if !errors.Is(err, jwtmanager.ErrNoPrivateKey) {
		t.Errorf("err = %v, want ErrNoPrivateKey", err)
	}
}

func TestGenerateWithoutPublicKey(t *testing.T) {
	privatePEM, _ := generateKeyPair(t)

	m, err := jwtmanager.NewManager(privatePEM, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	accessToken, err := m.Generate("user-1", "user", jwtmanager.TokenAccess, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = m.Verify(accessToken, jwtmanager.TokenAccess)
	if !errors.Is(err, jwtmanager.ErrNoPublicKey) {
		t.Errorf("err = %v, want ErrNoPublicKey", err)
	}
}
