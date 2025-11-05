package middleware


import (
	"net/http"
	"os"
	"context"


	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	userIDKey contextKey = "user_id"
	emailKey  contextKey = "email"
)

func AuthMiddleware(next http.Handler) http.Handler {
	secret := []byte(os.Getenv("JWT_SECRET"))
	
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenHeader := r.Header.Get("Authorization")

		if tokenHeader == "" || len(tokenHeader) < 8{
			http.Error(w,"Unauthorized", http.StatusUnauthorized)
			return;
		}

		tokenString := tokenHeader[len("Bearer "):]

		token,err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
			return secret,nil
		})

		if err != nil || !token.Valid {

			http.Error(w,"Unauthorized", http.StatusUnauthorized)
			return;

		}

		claims, ok := token.Claims.(jwt.MapClaims)

		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		userID, _ := claims["user_id"].(string)
		email, _ := claims["email"].(string)

		// ✅ Attach user info into context
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		ctx = context.WithValue(ctx, emailKey, email)

		// Pass the request forward with new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) (string, bool) {
    id, ok := ctx.Value(userIDKey).(string)
    return id, ok
}

func GetEmail(ctx context.Context) (string, bool) {
    email, ok := ctx.Value(emailKey).(string)
    return email, ok
}
