package internal

import "errors"

var (
	ErrConflictingResponse           = errors.New("conflicting response")
	ErrDailyLimitExceeded            = errors.New("daily response limit exceeded")
	ErrEmailAlreadyExists            = errors.New("email already exists")
	ErrInvalidCredentials            = errors.New("invalid credentials")
	ErrDailyInteractionLimitExceeded = errors.New("daily interaction limit exceeded")
	ErrFeatureNotFound               = errors.New("feature not found")
	ErrFeatureAlreadySubscribed      = errors.New("feature already subscribed")
	ErrUserNotFound                  = errors.New("user not found")
)
