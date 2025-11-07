package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	userIDKey contextKey = "user_id"
	emailKey  contextKey = "email"
)


// ------------------------------------------------------------
// STRICT AUTH MIDDLEWARE (Requires Login)
// ------------------------------------------------------------
func AuthMiddleware(next http.Handler) http.Handler {
	secret := []byte(os.Getenv("JWT_SECRET"))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
			return secret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		userID, _ := claims["user_id"].(string)
		email, _ := claims["email"].(string)

		// Attach to context
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		ctx = context.WithValue(ctx, emailKey, email)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}



// ------------------------------------------------------------
// OPTIONAL AUTH MIDDLEWARE (Public Routes + Logged-in Upgrade)
// ------------------------------------------------------------
func OptionalMiddleware(next http.Handler) http.Handler {
	secret := []byte(os.Getenv("JWT_SECRET"))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")

		// ✅ No token → proceed as anonymous
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			next.ServeHTTP(w, r)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
			return secret, nil
		})

		// ✅ Invalid token → treat as anonymous
		if err != nil || !token.Valid {
			next.ServeHTTP(w, r)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		userID, _ := claims["user_id"].(string)
		email, _ := claims["email"].(string)

		ctx := r.Context()

		// ✅ Attach user identity if present
		if userID != "" {
			ctx = context.WithValue(ctx, userIDKey, userID)
		}
		if email != "" {
			ctx = context.WithValue(ctx, emailKey, email)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}



// ------------------------------------------------------------
// HELPERS
// ------------------------------------------------------------
func GetUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}

func GetEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(emailKey).(string)
	return email, ok
}
