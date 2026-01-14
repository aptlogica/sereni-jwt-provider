package errors

import "errors"

var (
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrInvalidToken            = errors.New("invalid token")
	ErrInvalidRefreshTokenSvc  = errors.New("invalid refresh token")
	ErrNotRefreshToken         = errors.New("token is not a refresh token")
	ErrTokenUserMismatch       = errors.New("token user mismatch")

	// Repository level errors
	ErrUserExists          = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrTokenNotFound       = errors.New("token not found")
)
