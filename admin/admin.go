package admin

import (
	"context"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/spaceuptech/helpers"
)

// Module stores admin module information
type Module struct {
	lock   sync.RWMutex
	secret string
}

// New initializes admin module
func New(secret string) *Module {
	return &Module{secret: secret}
}

// CreateToken create jwt token
func (m *Module) CreateToken(ctx context.Context, tokenClaims map[string]interface{}) (string, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	claims := jwt.MapClaims{}
	for k, v := range tokenClaims {
		claims[k] = v
	}

	// Add expiry of one week
	claims["exp"] = time.Now().Add(10 * time.Minute).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = "sc-admin-kid"
	tokenString, err := token.SignedString([]byte(m.secret))
	if err != nil {
		return "", helpers.Logger.LogError(helpers.GetRequestID(ctx), "Cannot sign token with the given secret", err, nil)
	}

	return tokenString, nil
}
