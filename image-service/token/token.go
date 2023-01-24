package token

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bernardn38/socialsphere/image-service/helpers"
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
		// log.Printf("%+v", r)
		auth := r.Header.Get("Cookie")
		fields := strings.Split(auth, ";")
		var tokenString string
		for _, cookie := range fields {
			cookieFields := strings.Split(cookie, "=")
			if strings.Contains(cookieFields[0], "jwtToken") && len(cookieFields) > 1 {
				tokenString = cookieFields[1]
			}
		}
		token, err := tm.VerifyToken(tokenString)
		if err != nil {
			log.Println(err)
			helpers.ResponseNoPayload(w, 401)
			return
		}

		ctx := context.WithValue(r.Context(), "userId", token.ID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (tm *Manager) VerifyToken(token string) (*jwt.RegisteredClaims, error) {
	// create a Verifier (HMAC in this example)
	verifier, err := jwt.NewVerifierHS(tm.SigningMethod, tm.Secret)
	if err != nil {

		return nil, err
	}

	// parse and verify a token
	tokenBytes := []byte(token)
	newToken, err := jwt.Parse(tokenBytes, verifier)
	if err != nil {
		log.Println(err)

		return nil, err
	}

	// or just verify it's signature
	err = verifier.Verify(newToken)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// get Registered claims
	var newClaims jwt.RegisteredClaims
	errClaims := json.Unmarshal(newToken.Claims(), &newClaims)
	if errClaims != nil {
		log.Println(errClaims)
		return nil, err
	}

	// or parse only claims
	errParseClaims := jwt.ParseClaims(tokenBytes, verifier, &newClaims)
	if errParseClaims != nil {
		log.Println(errParseClaims)
		return nil, err
	}

	if newClaims.ExpiresAt.Before(time.Now()) {
		log.Println("Error: token expired")
		return nil, errors.New("token erpired")
	}

	return &newClaims, nil
}
