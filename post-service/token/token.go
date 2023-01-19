package token

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bernardn38/socialsphere/post-service/helpers"
	"github.com/cristalhq/jwt/v4"
)

type Manager struct {
	Secret        []byte
	SigningMethod jwt.Algorithm
	Verifier      jwt.Verifier
}

func NewManager(secret []byte, SigningMethod jwt.Algorithm) *Manager {
	return &Manager{
		Secret:        secret,
		SigningMethod: SigningMethod,
	}
}
func (tm *Manager) VerifyJwtToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Cookie")
		fields := strings.Split(auth, ";")
		var tokenString string
		for _, cookie := range fields {
			cookieFields := strings.Split(cookie, "=")
			if strings.Contains(cookieFields[0], "jwtToken") && len(cookieFields) > 1 {
				tokenString = cookieFields[1]
			}
		}
		token, ok := tm.VerifyToken(tokenString)
		if !ok {
			helpers.ResponseNoPayload(w, 401)
			return
		}
		ctx := context.WithValue(r.Context(), "userId", token.ID)
		ctx = context.WithValue(ctx, "username", token.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (tm *Manager) VerifyToken(token string) (*jwt.RegisteredClaims, bool) {
	// create a Verifier (HMAC in this example)
	verifier, err := jwt.NewVerifierHS(tm.SigningMethod, tm.Secret)
	if err != nil {
		return nil, false
	}

	// parse and verify a token
	tokenBytes := []byte(token)
	parsedToken, err := jwt.Parse(tokenBytes, verifier)
	if err != nil {
		log.Println(err, "parse")
		return nil, false
	}
	// or just verify it's signature
	err = verifier.Verify(parsedToken)
	if err != nil {
		log.Println(err, "signature")
		return nil, false
	}

	// get Registered claims
	var newClaims jwt.RegisteredClaims
	errClaims := json.Unmarshal(parsedToken.Claims(), &newClaims)
	if errClaims != nil {
		log.Println(errClaims, "claims")
		return nil, false
	}
	// or parse only claims
	errParseClaims := jwt.ParseClaims(tokenBytes, verifier, &newClaims)
	if errParseClaims != nil {
		log.Println(errParseClaims)
		return nil, false
	}
	if newClaims.ExpiresAt.Before(time.Now()) {
		log.Println("Error: token expired")
		return nil, false
	}

	return &newClaims, true
}
