package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)


func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tokenString := r.Header.Get("Authorization")
        if tokenString == "" {
            http.Error(w, "Missing token", http.StatusUnauthorized)
            return
        }

        tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
        token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
            // Verify the token signature using the shared secret or public key
            return []byte("your-secret-key"), nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        claims := token.Claims.(*jwt.StandardClaims)
        userID := claims.Subject
        userRole := claims.Audience

        // Pass the user information to the next handler
        ctx := context.WithValue(r.Context(), "userID", userID)
        ctx = context.WithValue(ctx, "userRole", userRole)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

