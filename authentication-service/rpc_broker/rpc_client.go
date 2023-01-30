package rpcbroker

import (
	"errors"
	"net/rpc"
)

type RpcClient struct {
}
type CreateUserParams struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	UserId    int32  `json:"userId"`
	Username  string `json:"username"`
	Email     string `json:"email"`
}

func (r *RpcClient) CreateFriendServiceUser(createUserParams CreateUserParams) error {
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

func (r *RpcClient) CreateIdentityServiceUser(createUserParams CreateUserParams) error {
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
