package service

import (
	"strconv"
)

type PaginationForm struct {
	PageSize int64 `json:"pageSize" validate:"required,min=1,max=20"`
	PageNo   int64 `json:"pageNo" validate:"required,min=1"`
}

func ValidatePagination(pageSize string, pageNo string) (int32, int32) {

	parsedPageSize, err := strconv.ParseInt(pageSize, 10, 64)
	if err != nil {
		parsedPageSize = 10
	}
	parsedPageNo, err := strconv.ParseInt(pageNo, 10, 64)
	if err != nil {
		parsedPageNo = 1
	}

	offset := (parsedPageNo - 1) * parsedPageSize
	return int32(parsedPageSize), int32(offset)
}
