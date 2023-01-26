package rpcreceiver

import (
	"context"
	"log"
	"net"
	"net/rpc"

	"github.com/bernardn38/socialsphere/friend-service/sql/users"
)

type RpcReceiver struct {
	FriendService *FriendService
}
type CreateUserParams struct {
	FirstName string
	LastName  string
	UserId    int32
	Username  string
	Email     string
}

type FriendService struct {
	UserDb *users.Queries
}

func NewRpcReceiver(userDb *users.Queries) *RpcReceiver {
	rpcReceiver := RpcReceiver{}
	IdentityService := FriendService{UserDb: userDb}
	rpcReceiver.FriendService = &IdentityService
	return &rpcReceiver
}

func (s *FriendService) CreateUser(user CreateUserParams, reply *bool) error {
	_, err := s.UserDb.CreateUser(context.Background(), users.CreateUserParams{
		UserID:    user.UserId,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	})
	if err != nil {
		log.Println(err)
		*reply = false
		return err
	}
	*reply = true
	return nil
}

func (r *RpcReceiver) ListenForRpc() {
	server := rpc.NewServer()
	server.Register(r.FriendService)

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
