package rpcemitter

import (
	"errors"
	"net/rpc"
)

type RpcEmitter struct {
}
type CreateUserParams struct {
	FirstName string
	LastName  string
	UserId    int32
	Username  string
	Email     string
}

func (r *RpcEmitter) CreateFriendServiceUser(createUserParams CreateUserParams) error {
	var reply bool
	friendServiceConnection, err := rpc.Dial("tcp", "friend-service:9002")
	if err != nil {
		return err
	}
	defer friendServiceConnection.Close()
	err = friendServiceConnection.Call("FriendService.CreateUser", createUserParams, &reply)
	if err != nil {
		return err
	}
	if reply == false {
		return errors.New("error registering user in friend service")
	}
	return nil
}

func (r *RpcEmitter) CreateIdentityServiceUser(createUserParams CreateUserParams) error {
	var reply bool
	friendServiceConnection, err := rpc.Dial("tcp", "identity-service:9002")
	if err != nil {
		return err
	}
	defer friendServiceConnection.Close()
	err = friendServiceConnection.Call("IdentityService.CreateUser", createUserParams, &reply)
	if err != nil {
		return err
	}
	if reply == false {
		return errors.New("error registering user in identity service")
	}
	return nil
}
