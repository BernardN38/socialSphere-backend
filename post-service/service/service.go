package service

import (
	"context"
	"database/sql"
	"log"
	"mime/multipart"
	"time"

	"github.com/bernardn38/socialsphere/post-service/models"
	"github.com/bernardn38/socialsphere/post-service/rabbitmq_broker"
	rpcbroker "github.com/bernardn38/socialsphere/post-service/rpc_broker"
	"github.com/bernardn38/socialsphere/post-service/sql/post"
	"github.com/minio/minio-go"
)

type PostService struct {
	PostDb          *post.Queries
	RabbitMQEmitter *rabbitmq_broker.RabbitMQEmitter
	RpcClient       *rpcbroker.RpcClient
	MinioClient     *minio.Client
}

func New(config *models.Config) (*PostService, error) {
	//open connection to postgres
	db, err := sql.Open("postgres", config.PostgresUrl)
	if err != nil {
		return nil, err
	}

	// init sqlc user queries
	queries := post.New(db)

	//init rabbitmq message emitter
	rabbitMQConn := rabbitmq_broker.ConnectToRabbitMQ(config.RabbitmqUrl)
	rabbitBroker, err := rabbitmq_broker.NewRabbitEventEmitter(rabbitMQConn)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	rabbitmq_broker.NewRabbitEventEmitter(rabbitMQConn)

	minioClient, err := minio.New("minio:9000", config.MinioKey, config.MinioSecret, false)
	if err != nil {
		return nil, err
	}
	return &PostService{
		PostDb:          queries,
		RabbitMQEmitter: &rabbitBroker,
		RpcClient:       &rpcbroker.RpcClient{},
		MinioClient:     minioClient,
	}, nil
}

func (s *PostService) GetLikeCuountbyPostId(postId int32) (int64, error) {
	likeCount, err := s.PostDb.GetPostLikeCountById(context.Background(), postId)
	if err != nil {
		return 0, err
	}
	return likeCount, nil
}

func (s *PostService) CreateCommentForPostId(commentForm models.CreateCommentForm, postId int32, userId int32, username string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	createdComment, err := s.PostDb.CreateComment(ctx, post.CreateCommentParams{Body: commentForm.Body, UserID: userId, AuthorName: username})
	if err != nil {
		return err
	}

	_, err = s.PostDb.CreatePostComment(context.Background(), post.CreatePostCommentParams{
		PostID:    postId,
		CommentID: createdComment.ID,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *PostService) GetCommentByPostId(postId int32) ([]post.GetAllPostCommentsByPostIdRow, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	postsComments, err := s.PostDb.GetAllPostCommentsByPostId(ctx, postId)
	if err != nil {
		return nil, err
	}
	return postsComments, nil
}

func (s *PostService) CreatPost(postForm models.CreatPostForm, file multipart.File, header *multipart.FileHeader) (post.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if file != nil {
		_, err := s.MinioClient.PutObject("media-service-socialsphere1", postForm.ImageID.UUID.String(), file, header.Size, minio.PutObjectOptions{ContentType: header.Header.Get("Content-Type")})
		if err != nil {
			return post.Post{}, err
		}
		file.Close()
	}
	createdPost, err := s.PostDb.CreatePost(ctx, post.CreatePostParams{
		Body:       postForm.Body,
		UserID:     postForm.UserID,
		AuthorName: postForm.AuthorName,
		ImageID:    postForm.ImageID,
	})
	if err != nil {
		return post.Post{}, err
	}
	return createdPost, nil
}

func (s *PostService) GetPostsByUserIdPaginated(userId int32, pageNo string, pageSize string) (models.PostPage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	limit, offset := ValidatePagination(pageSize, pageNo)
	lastPage := false
	posts, err := s.PostDb.GetPostByUserIdPaged(ctx, post.GetPostByUserIdPagedParams{
		UserID: userId,
		Limit:  limit + 1,
		Offset: offset,
	})
	if err != nil {
		return models.PostPage{}, err
	}
	if len(posts) > int(limit) {
		lastPage = false
		posts = posts[:limit]
	} else {
		lastPage = true
	}
	return models.PostPage{Posts: posts,
		PageSize: len(posts),
		PageNo:   (offset / limit) + 1,
		LastPage: lastPage}, nil
}

func (s *PostService) GetPostWithLikes(postId int32) (post.GetPostByIdWithLikesRow, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	respPost, err := s.PostDb.GetPostByIdWithLikes(ctx, postId)
	if err != nil {
		return post.GetPostByIdWithLikesRow{}, err
	}
	return respPost, nil
}

func (s *PostService) DeletePost(postId int32, userId int32) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	imageId, err := s.PostDb.DeletePostById(ctx, post.DeletePostByIdParams{
		ID:     postId,
		UserID: userId,
	})
	if err != nil {
		return err
	}
	err = s.RabbitMQEmitter.PushDelete(imageId.UUID.String())
	if err != nil {
		log.Println(err)
	}
	return nil
}

func (s *PostService) CreatePostLike(postId int32, userId int32) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, err := s.PostDb.CreatePostLike(ctx, post.CreatePostLikeParams{
		PostID: postId,
		UserID: userId,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *PostService) DeletePostLike(postId int32, userId int32) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err := s.PostDb.DeletePostLike(ctx, post.DeletePostLikeParams{
		PostID: postId,
		UserID: userId,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *PostService) CheckLike(postId int32, userId int32) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	isLiked, err := s.PostDb.CheckLike(ctx, post.CheckLikeParams{
		PostID: postId,
		UserID: userId,
	})
	return isLiked, err
}
