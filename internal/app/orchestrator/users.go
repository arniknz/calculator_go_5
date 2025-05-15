package application

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtkey = []byte(os.Getenv("JWT_KEY"))

type Assert struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

type user_id_key struct{}

func HashPassword(s string) (string, error) {
	saltedBytes := []byte(s)
	hashedBytes, err := bcrypt.GenerateFromPassword(saltedBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	hash := string(hashedBytes[:])
	return hash, nil
}

func CheckPassword(hash, s string) error {
	pwd := []byte(s)
	pwd_from_db := []byte(hash)
	return bcrypt.CompareHashAndPassword(pwd_from_db, pwd)
}

func (o *Orchestrator) AuthMiddleware(request http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(w, "error: missing token", http.StatusUnauthorized)
			return
		}

		assert := &Assert{}
		token, err := jwt.ParseWithClaims(tokenStr, assert, func(t *jwt.Token) (interface{}, error) {
			return jwtkey, nil
		},
		)

		if err != nil || !token.Valid {
			http.Error(w, "error: invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), user_id_key{}, assert.UserID)
		request.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CreateToken(user_id int) (string, error) {
	assert := &Assert{
		UserID: user_id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, assert)
	return token.SignedString(jwtkey)
}
