package rpc_broker

import (
	"bytes"
	"context"
	"io"
	"log"
	"net"
	"net/rpc"

	"github.com/bernardn38/socialsphere/image-service/models"
	"github.com/bernardn38/socialsphere/image-service/sql/userImages"
	"github.com/google/uuid"

	"github.com/minio/minio-go"
)

type RpcServer struct {
	ImageService *MediaService
}
type CreateUserParams struct {
	FirstName string
	LastName  string
	UserId    int32
	Username  string
	Email     string
}

type MediaService struct {
	ImageDb     *userImages.Queries
	MinioClient *minio.Client
}

func NewRpcServer(imageDb *userImages.Queries, minioClient *minio.Client) *RpcServer {
	rpcReceiver := RpcServer{}
	MediaService := MediaService{ImageDb: imageDb, MinioClient: minioClient}
	rpcReceiver.ImageService = &MediaService
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

func (s *MediaService) UploadImage(imageUpload models.RpcImageUpload, reply *bool) error {
	_, err := s.ImageDb.CreateImage(context.Background(), userImages.CreateImageParams{
		UserID:  imageUpload.UserId,
		ImageID: imageUpload.ImageId,
	})
	if err != nil {
		log.Println(err)
		*reply = false
		return err
	}
	buf := bytes.NewBuffer(imageUpload.Image)
	_, err = s.MinioClient.PutObject("media-service-socialsphere1", imageUpload.ImageId.String(), buf, int64(len(imageUpload.Image)), minio.PutObjectOptions{ContentType: imageUpload.ContentType})
	if err != nil {
		log.Println(err)
		*reply = false
		return err
	}
	*reply = true
	return nil
}

func (s *MediaService) GetImage(imageid uuid.UUID, reply *[]byte) error {
	object, err := s.MinioClient.GetObject("media-service-socialsphere1", imageid.String(), minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)
	io.Copy(buf, object)
	*reply = buf.Bytes()
	return nil
}
