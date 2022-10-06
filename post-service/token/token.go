package token

import (
	"context"
	"encoding/json"
	"github.com/bernardn38/socialsphere/post-service/helpers"
	"github.com/cristalhq/jwt/v4"
	"log"
	"net/http"
	"strings"
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
		tokenString := strings.TrimPrefix(auth, "jwtToken=")
		token, ok := tm.VerifyToken(tokenString)
		if !ok {
			helpers.ResponseNoPayload(w, 401)
			return
		}
		log.Println(token)
		ctx := context.WithValue(r.Context(), "token", token)

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
	newToken, err := jwt.Parse(tokenBytes, verifier)
	if err != nil {
		log.Println(err)

		return nil, false
	}

	// or just verify it's signature
	err = verifier.Verify(newToken)
	if err != nil {
		log.Println(err)
		return nil, false
	}

	// get Registered claims
	var newClaims jwt.RegisteredClaims
	errClaims := json.Unmarshal(newToken.Claims(), &newClaims)
	if errClaims != nil {
		log.Println(errClaims)
		return nil, false
	}

	// or parse only claims
	errParseClaims := jwt.ParseClaims(tokenBytes, verifier, &newClaims)
	if errParseClaims != nil {
		log.Println(errParseClaims)
		return nil, false
	}

	return &newClaims, true
}
