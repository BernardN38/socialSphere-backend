package rpcbroker

import (
	"context"
	"log"
	"net"
	"net/rpc"

	"github.com/bernardn38/socialsphere/identity-service/sql/users"
)

type RpcServer struct {
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

func RunRpcServer(userDb users.Queries) {
	//listen for calls over rpc
	rpcReceiver := NewRpcServer(&userDb)
	go rpcReceiver.ListenForRpc()
}
func NewRpcServer(userDb *users.Queries) *RpcServer {
	rpcReceiver := RpcServer{}
	IdentityService := IdentityService{UserDb: userDb}
	rpcReceiver.IdentityService = &IdentityService
	return &rpcReceiver
}
func (s *IdentityService) CreateUser(user *CreateUserParams, reply *bool) error {
	log.Println("creating identity service user via rpc")
	err := s.UserDb.CreateUser(context.Background(), users.CreateUserParams{
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

func (r *RpcServer) ListenForRpc() {
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
