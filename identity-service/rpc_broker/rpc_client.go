package rpcbroker

import (
	"errors"
	"log"
	"net/rpc"

	"github.com/bernardn38/socialsphere/identity-service/models"
	"github.com/google/uuid"
)

type RpcClient struct {
}

func (r *RpcClient) UploadImage(imageUpload models.RpcImageUpload) error {
	var reply bool
	mediaServiceConnection, err := rpc.Dial("tcp", "media-service:9002")
	if err != nil {
		log.Println("err making image servive connection", err)
		return err
	}
	// defer imageServiceConnection.Close()
	err = mediaServiceConnection.Call("MediaService.UploadImage", imageUpload, &reply)
	if err != nil {
		log.Println("err makeing call to media service upload image", err)
		return err
	}
	if !reply {
		return errors.New("error registering user in friend service")
	}
	return nil
}

func (r *RpcClient) GetImage(imageId uuid.UUID) ([]byte, error) {
	var reply []byte
	mediaServiceConnection, err := rpc.Dial("tcp", "media-service:9002")
	if err != nil {
		log.Println("err making image servive connection", err)
		return nil, err
	}
	// defer imageServiceConnection.Close()
	err = mediaServiceConnection.Call("MediaService.GetImage", imageId, &reply)
	if err != nil {
		log.Println("err makeing call to media service upload image", err)
		return nil, err
	}
	if reply == nil {
		return nil, errors.New("error registering user in friend service")
	}
	return reply, nil
}
