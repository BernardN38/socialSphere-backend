package models

import (
	"time"

	"github.com/bernardn38/socialsphere/post-service/sql/post"
	"github.com/google/uuid"
)

type Post struct {
	Body       string    `json:"body" validate:"required"`
	Author     int       `json:"author" validate:"required"`
	AuthorName string    `json:"authorName" validate:"required"`
	CreatedAt  time.Time `json:"created_at"`
}

type CommentsResp struct {
	Body      string    `json:"body"`
	CommentId uuid.UUID `json:"comment_id"`
}
type PostLikes struct {
	PostId    string `json:"postId"`
	LikeCount int64  `json:"likeCount"`
}

type RpcImageUpload struct {
	UserId  int32
	Image   []byte
	ImageId uuid.UUID
}

type PostPage struct {
	Posts    []post.GetPostByUserIdPagedRow `json:"posts"`
	PageNo   int32                          `json:"pageNo"`
	PageSize int                            `json:"pageSize"`
	LastPage bool                           `json:"lastPage"`
}
