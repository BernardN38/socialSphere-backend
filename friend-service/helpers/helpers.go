package helpers

import (
	"net/http"
)

type UserView struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
type RegisterForm struct {
	Username  string `json:"username" validate:"required,min=2,max=100"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8,max=128"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
}

func ResponseWithPayload(w http.ResponseWriter, responseCode int, payload []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	_, _ = w.Write(payload)
}
func ResponseNoPayload(w http.ResponseWriter, responseCode int) {
	w.WriteHeader(responseCode)
}
