package jwtmanager

import "errors"

var (
	ErrNoPrivateKey     = errors.New("private key is not set")
	ErrNoPublicKey      = errors.New("public key is not set")
	ErrInvalidToken     = errors.New("invalid token")
	ErrWrongTokenType   = errors.New("wrong token type")
	ErrTokenExpired     = errors.New("token expired")
	ErrInvalidSignature = errors.New("invalid token signature")
)
