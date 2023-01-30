package rpcbroker

import (
	"errors"
	"log"
	"net/rpc"

	"github.com/bernardn38/socialsphere/post-service/models"
)

type RpcClient struct {
}

func (r *RpcClient) UploadImage(imageUpload models.RpcImageUpload) error {
	var reply bool
	imageServiceConnection, err := rpc.Dial("tcp", "image-service:9002")
	if err != nil {
		log.Println("err making image servive connection", err)
		return err
	}
	// defer imageServiceConnection.Close()
	err = imageServiceConnection.Call("ImageService.UploadImage", imageUpload, &reply)
	if err != nil {
		log.Println("err makeing call to image service upload image", err)
		return err
	}
	if reply == false {
		return errors.New("error registering user in friend service")
	}
	return nil
}
