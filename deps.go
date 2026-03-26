package main

// This file ensures future dependencies are tracked in go.mod.
// These imports will be used when auth and rate limiting are implemented.
import (
	_ "github.com/golang-jwt/jwt/v5"
	_ "golang.org/x/crypto/bcrypt"
	_ "golang.org/x/oauth2"
	_ "golang.org/x/time/rate"
)
