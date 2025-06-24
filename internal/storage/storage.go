package storage

import "errors"

var (
	KeyNotFound   = errors.New("key not found")
	DuplicatedKey = errors.New("duplicated key")
)
