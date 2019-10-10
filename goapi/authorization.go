package goapi

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

func randomSecret() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return ``, err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// AuthMiddleware is middleware of JWT-based authorization
// - Header: "Authorization: Bearer <token>""
// - The JWT claim should include '{"id":"user@exampl.com"}'
// - the id will be compared against the 'auth' entries in the config
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, `No Authorization Header`, http.StatusUnauthorized)
			return
		}
		authParts := strings.Fields(auth)
		if len(authParts) != 2 || strings.ToLower(authParts[0]) != "bearer" {
			http.Error(w, `Invalid Authorization Header`, http.StatusUnauthorized)
			return
		}
		tokenString := authParts[1]
		claims, err := verifyToken(tokenString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if id, ok := claims["id"].(string); ok {
			if _, ok := config.Auths[id]; !ok {
				http.Error(w, fmt.Sprintf(`id %s is not authorized`, id), http.StatusUnauthorized)
				return
			}
			// authorization successful
			r.Header.Set("userID", id)
			next.ServeHTTP(w, r)
			return
		}
		http.Error(w, `unauthorized`, http.StatusUnauthorized)
	})
}

func verifyToken(tokenString string) (jwt.MapClaims, error) {
	// callback useb by jwt.Parse to get secret key based on token
	getKey := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Secret), nil
	}
	token, err := jwt.Parse(tokenString, getKey)
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}
