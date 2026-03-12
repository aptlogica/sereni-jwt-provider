// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package errors

import "errors"

var (
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrInvalidToken            = errors.New("invalid token")
	ErrInvalidRefreshTokenSvc  = errors.New("invalid refresh token")
	ErrNotRefreshToken         = errors.New("token is not a refresh token")
	ErrTokenUserMismatch       = errors.New("token user mismatch")
	ErrTokenExpire             = errors.New("token has expired")

	// Repository level errors
	ErrUserExists          = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrTokenNotFound       = errors.New("token not found")
)
