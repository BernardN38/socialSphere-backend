package handler

import (
	"encoding/json"
	"gopkg.in/go-playground/validator.v9"
)

func ValidatePostForm(reqBody []byte) (*Post, error) {
	var form Post
	err := json.Unmarshal(reqBody, &form)
	if err != nil {
		return nil, err
	}

	v := validator.New()
	err = v.Struct(form)
	if err != nil {
		return nil, err
	}

	return &form, nil
}
