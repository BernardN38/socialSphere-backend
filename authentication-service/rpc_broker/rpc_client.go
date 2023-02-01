package rpcbroker

import (
	"errors"
	"net/rpc"

	"github.com/bernardn38/socialsphere/authentication-service/models"
)

type RpcClient struct {
}

func (r *RpcClient) CreateFriendServiceUser(createUserParams models.CreateUserParams) error {
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

func (r *RpcClient) CreateIdentityServiceUser(createUserParams models.CreateUserParams) error {
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
