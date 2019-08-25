package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, `unauthorized`, http.StatusUnauthorized)
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
			http.Error(w, `unauthorized`, http.StatusUnauthorized)
			return
		}
		userID := fmt.Sprintf("%v", claims["user_id"])
		r.Header.Set("userID", userID)
		next.ServeHTTP(w, r)
	})
}

func verifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Secret), nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}
