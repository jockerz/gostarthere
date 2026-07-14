package tasks

import "errors"

const (
	TypeAuthEmail       = "email:auth"
	TypeUserEmailUpdate = "email:user:confirm_email_update"
)

var (
	ErrTokenActionCreateTask  = errors.New("Fail to create task")
	ErrTokenActionEnqueueTask = errors.New("Fail to enqueue task")
)
