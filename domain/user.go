package domain

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

func NewUser(ctx context.Context, name, password string, isAdmin bool) *User {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		// handle error
		log.Fatalf("Failed to hash password: %v", err)
	}

	return &User{
		// ID:           idgen.New("usr"),
		Name:         name,
		PasswordHash: hashedPassword,
		Created:      time.Now(),
		IsAdmin:      isAdmin,
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
    err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password))
    return err == nil
}

func (u *User) Session() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": u.ID,
	})

	return token.SignedString([]byte(os.Getenv("TOKEN_SECRET")))
}

func hashPassword(pw string) ([]byte, error) {
    hashed, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }
    return hashed, nil
}
