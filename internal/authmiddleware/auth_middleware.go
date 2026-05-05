package authmiddleware

import (
	"context"
	"net/http"

	"github.com/carloscfgos1980/ecom-api/internal/utils"
)

// HTTP middleware setting a value on the request context
func AuthMiddleware(next http.Handler, jwtSecret string) http.Handler {
	// Return a new http.HandlerFunc that wraps the original handler and adds the authentication logic
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the token from the Authorization header
		token, err := utils.GetBearerToken(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		// Validate the token and extract the customer ID
		customerID, err := utils.ValidateJWT(token, jwtSecret)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		// Create a new context with the customer ID value
		ctx := context.WithValue(r.Context(), "customerID", customerID)

		// Call the next handler with the new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
