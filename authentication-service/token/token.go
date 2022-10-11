package token

import (
	"context"
	"encoding/json"
	"github.com/bernardn38/socialsphere/authentication-service/helpers"
	"github.com/cristalhq/jwt/v4"
	"log"
	"net/http"
	"strings"
	"time"
)

type Manager struct {
	Secret        []byte
	SigningMethod jwt.Algorithm
	Verifier      jwt.Verifier
}

type tokenManger interface {
	func() jwt.Token
}

func NewManager(secret []byte, SigningMethod jwt.Algorithm) *Manager {
	return &Manager{
		Secret:        secret,
		SigningMethod: SigningMethod,
	}
}
func (tm *Manager) GenerateToken(userId string, TTL time.Duration) (*jwt.Token, error) {
	signer, err := jwt.NewSignerHS(tm.SigningMethod, tm.Secret)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// create claims (you can create your own, see: Example_BuildUserClaims)
	claims := &jwt.RegisteredClaims{
		ID:        userId,
		ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(TTL)},
		IssuedAt:  &jwt.NumericDate{Time: time.Now()},
	}

	// create a Builder
	builder := jwt.NewBuilder(signer)

	// and build a Token
	token, err := builder.Build(claims)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return token, nil
}

func (tm *Manager) VerifyJwtToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		auth := r.Header.Get("Authorization")
		tokenString := strings.TrimPrefix(auth, "Bearer ")
		token, ok := tm.VerifyToken(tokenString)
		if !ok {
			helpers.ResponseNoPayload(w, 401)
			return
		}

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
