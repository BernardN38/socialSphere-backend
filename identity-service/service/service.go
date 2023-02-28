package service

import (
	"bytes"
	"context"
	"io"
	"log"
	"mime/multipart"
	"time"

	"github.com/bernardn38/socialsphere/identity-service/models"
	"github.com/bernardn38/socialsphere/identity-service/rabbitmq_broker"
	rpcbroker "github.com/bernardn38/socialsphere/identity-service/rpc_broker"
	"github.com/bernardn38/socialsphere/identity-service/sql/users"
	"github.com/google/uuid"
)

type IdentityService struct {
	UserDb        *users.Queries
	RabbitEmitter *rabbitmq_broker.RabbitBroker
	RpcClient     *rpcbroker.RpcClient
}

func New(userDb *users.Queries, rabbitEmitter *rabbitmq_broker.RabbitBroker, rpcClient *rpcbroker.RpcClient) (*IdentityService, error) {
	return &IdentityService{
		UserDb:        userDb,
		RabbitEmitter: rabbitEmitter,
		RpcClient:     &rpcbroker.RpcClient{},
	}, nil
}

func (s *IdentityService) CreateUser(user models.UserForm) error {
	//create new user in database
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err := s.UserDb.CreateUser(ctx, users.CreateUserParams{ID: user.UserId,
		Username: user.Username, FirstName: user.Username, LastName: user.LastName, Email: user.Email})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s *IdentityService) CreateUserProfileImage(userId int32, file multipart.File, contentType string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	// create profile image association in database
	imageId := uuid.New()
	err := s.UserDb.CreateUserProfileImage(ctx, users.CreateUserProfileImageParams{
		UserID:  userId,
		ImageID: imageId,
	})
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	if err != nil {
		return err
	}
	//send image to rabbitmq for processing and upload to s3 bucket
	imageUpload := models.RpcImageUpload{
		UserId:  userId,
		Image:   buf.Bytes(),
		ImageId: imageId,
	}
	err = s.RpcClient.UploadImage(imageUpload)
	if err != nil {
		log.Println("rpc error", err)
		return err
	}
	if file != nil {
		err = SendImageToQueue(s, "image-proccessing", imageId, contentType)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *IdentityService) GetUser(userId int32) (users.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	//get from database
	user, err := s.UserDb.GetUserById(ctx, userId)
	if err != nil {
		return users.User{}, err
	}
	return user, nil
}

func (s *IdentityService) GetUserProfileImage(userId int32) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	//get image id from datbase for specified user id
	imageId, err := s.UserDb.GetUserProfileImage(ctx, userId)
	if err != nil {
		return nil, err
	}

	imageBytes, err := s.RpcClient.GetImage(imageId)
	if err != nil {
		return nil, err
	}
	return imageBytes, nil
}

func (s *IdentityService) GetImageIdByUserId(userId int32) (uuid.UUID, error) {
	imageId, err := s.UserDb.GetUserProfileImage(context.Background(), userId)
	if err != nil {
		return uuid.UUID{}, err
	}
	return imageId, nil
}
