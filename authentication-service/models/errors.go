package models

import "fmt"

type RpcCreateUserError struct {
	IdentityServiceError error
	FriendServiceError   error
}

func (r RpcCreateUserError) Error() string {
	return fmt.Sprintf("%v, %v", r.IdentityServiceError.Error(), r.FriendServiceError.Error())
}
