package common

import "errors"

var (
	ErrLockAlreadyRequired = errors.New("锁已经被占用")
)
