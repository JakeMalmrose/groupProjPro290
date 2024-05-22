package authmiddleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

func Authorize(next http.Handler, allowedRoles ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			// Verify the token signature using the shared secret or public key
			return []byte("zachariah-hansen"), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(*Claims)
        fmt.Printf("Claims: %+v\n", claims)
		userID := claims.UserID
		fmt.Println("UserID:", userID)
		userRole := claims.Audience
        fmt.Println("UserRole:", userRole)
		if(allowedRoles == nil) {
			
		} else if !contains(allowedRoles, userRole) {
			log.Print("Allowedroles:", allowedRoles)
			log.Print("Userrole:", userRole)
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
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