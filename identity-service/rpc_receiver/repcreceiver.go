package rpcreceiver

import (
	"context"
	"net"
	"net/rpc"

	"github.com/bernardn38/socialsphere/identity-service/sql/users"
)

type RpcReceiver struct {
	IdentityService *IdentityService
}
type CreateUserParams struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	UserId    int32  `json:"userId"`
	Username  string `json:"username"`
	Email     string `json:"email"`
}
type IdentityService struct {
	UserDb *users.Queries
}

func NewRpcReceiver(userDb *users.Queries) *RpcReceiver {
	rpcReceiver := RpcReceiver{}
	IdentityService := IdentityService{UserDb: userDb}
	rpcReceiver.IdentityService = &IdentityService
	return &rpcReceiver
}
func (s *IdentityService) CreateUser(user *CreateUserParams, reply *bool) error {
	_, err := s.UserDb.CreateUser(context.Background(), users.CreateUserParams{
		ID:        user.UserId,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	})
	if err != nil {
		*reply = false
		return err
	}
	*reply = true
	return nil
}

func (r *RpcReceiver) ListenForRpc() {
	server := rpc.NewServer()
	server.Register(r.IdentityService)

	listener, err := net.Listen("tcp", ":9002")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go server.ServeConn(conn)
	}
}
