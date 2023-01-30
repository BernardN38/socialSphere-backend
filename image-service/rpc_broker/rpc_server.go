package rpc_broker

import (
	"context"
	"log"
	"net"
	"net/rpc"

	"github.com/bernardn38/socialsphere/image-service/helpers"
	"github.com/bernardn38/socialsphere/image-service/models"
	"github.com/bernardn38/socialsphere/image-service/sql/userImages"

	"github.com/minio/minio-go"
)

type RpcServer struct {
	ImageService *ImageService
}
type CreateUserParams struct {
	FirstName string
	LastName  string
	UserId    int32
	Username  string
	Email     string
}

type ImageService struct {
	ImageDb     *userImages.Queries
	MinioClient *minio.Client
}

func NewRpcServer(imageDb *userImages.Queries, minioClient *minio.Client) *RpcServer {
	rpcReceiver := RpcServer{}
	ImageService := ImageService{ImageDb: imageDb, MinioClient: minioClient}
	rpcReceiver.ImageService = &ImageService
	return &rpcReceiver
}
func RunRpcServer(imageDb *userImages.Queries, minioClient *minio.Client) {
	//listen for calls over rpc
	rpcReceiver := NewRpcServer(imageDb, minioClient)
	go rpcReceiver.ListenForRpc()
}
func (r *RpcServer) ListenForRpc() {
	server := rpc.NewServer()
	server.Register(r.ImageService)

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

func (s *ImageService) UploadImage(imageUpload models.RpcImageUpload, reply *bool) error {
	_, err := s.ImageDb.CreateImage(context.Background(), userImages.CreateImageParams{
		UserID:  imageUpload.UserId,
		ImageID: imageUpload.ImageId,
	})
	if err != nil {
		log.Println(err)
		*reply = false
		return err
	}
	err = helpers.UploadToS3(s.MinioClient, imageUpload.Image, imageUpload.ImageId.String())
	if err != nil {
		log.Println(err)
		*reply = false
		return err
	}
	*reply = true
	return nil
}
