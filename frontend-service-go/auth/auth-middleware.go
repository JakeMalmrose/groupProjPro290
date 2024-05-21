package authmiddleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)


func Authorize(next http.Handler, unauthorized http.Handler, allowedRoles ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			unauthorized.ServeHTTP(w, r)
			return;
		}

		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
		token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Verify the token signature using the shared secret or public key
			return []byte("zachariah-hansen"), nil
		})

		if err != nil || !token.Valid {
			unauthorized.ServeHTTP(w, r)
			return;
		}

		claims := token.Claims.(*jwt.StandardClaims)
		userID := claims.Subject
		userRole := claims.Audience

		if !contains(allowedRoles, userRole) {
			unauthorized.ServeHTTP(w, r)
			return;
		}

		// Pass the user information to the next handler
		ctx := context.WithValue(r.Context(), "userID", userID)
		ctx = context.WithValue(ctx, "userRole", userRole)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func contains(s []string, str string) bool {
    for _, v := range s {
        if v == str {
            return true
        }
    }
    return false
}
