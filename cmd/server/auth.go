package main

import (
	"errors"
	"net/http"
	"os"
	// "strings"

	"github.com/golang-jwt/jwt"

	"github.com/moov-io/base/log"	
)

func withAuth(logger log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve cookie from request
		tokenCookie, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				logger.LogErrorf("No cookie found")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}			
			logger.LogErrorf("Error occured while reading cookie")
			http.Error(w, "Bad Request", http.StatusBadRequest)			
			return
		}
		// Get the cookie value
		cookieValue := tokenCookie.Value

		// Log the cookie value
		logger.Logf("Token cookie value: %s", cookieValue)

		// tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		// if tokenString == "" {
		// 	logger.LogErrorf("Request missing token from %s to %s", r.RemoteAddr, r.URL.Path)
		// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
		// 	return
		// }

		secret := os.Getenv("API_KEY")
		
		// TEMP! INSECURE!
		logger.Logf("Secret: %s", secret)
		logger.Logf("Secret bytes:", []byte(secret))

		// Parse the jwt
		token, err := jwt.Parse(cookieValue, func(token *jwt.Token) (interface{}, error) {
			// Validate the algorithm used to sign the JWT
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}
			// Return the secret
			return []byte(secret), nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				logger.LogErrorf("Request with invalid signature from %s to %s", r.RemoteAddr, r.URL.Path)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			logger.LogErrorf("Request error from %s to %s: %v", r.RemoteAddr, r.URL.Path, err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if the JWT is valid
		if !token.Valid {
			logger.LogErrorf("Request with invalid token from %s to %s", r.RemoteAddr, r.URL.Path)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
