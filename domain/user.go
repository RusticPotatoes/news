package domain

import (
	"bytes"
	"context"
	"crypto/sha256"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/RusticPotatoes/news/idgen"
)

func NewUser(ctx context.Context, name, password string) *User {
	return &User{
		ID:           idgen.New("usr"),
		Name:         name,
		PasswordHash: hashPassword(password),
		Created:      time.Now(),
	}
}

type contextKey string

const userKey contextKey = "user"

func WithUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func UserFromContext(ctx context.Context) *User {
	u, ok := ctx.Value(userKey).(*User)
	if !ok {
		return nil
	}
	return u
}

type User struct {
	ID           string
	Name         string
	Created      time.Time
	PasswordHash []byte `json:"-"`
	IsAdmin      bool
}

func (u *User) ValidatePassword(password string) bool {
	if u == nil {
		return false
	}
	hash := hashPassword(password)
	return bytes.Equal(hash, u.PasswordHash)
}

func (u *User) Session() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": u.ID,
	})

	return token.SignedString([]byte(os.Getenv("TOKEN_SECRET")))
}

func hashPassword(pw string) []byte {
	h := sha256.New()
	h.Write([]byte(os.Getenv("PW_SALT")))
	h.Write([]byte(pw))
	hashed := h.Sum(nil)
	return hashed
}
