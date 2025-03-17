package domain

import "errors"

var (
	ErrAuthAttemptNotFound   = errors.New("auth attempt not found")
	ErrAuthAttemptNotPending = errors.New("auth attempt is not pending")
	ErrAuthLinkExpired       = errors.New("auth link expired")
)
