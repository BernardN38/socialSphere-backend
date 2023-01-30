package models

type CreateUserForm struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	UserId    int32  `json:"userId"`
	Username  string `json:"username"`
	Email     string `json:"email"`
}
