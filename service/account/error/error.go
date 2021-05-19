package error

import "github.com/1412335/grpc-rest-microservice/pkg/errors"

var (
	ErrMissingUserID = errors.BadRequest("Missing user id", map[string]string{"id": "Missing user id"})

	ErrUserNotFound  = errors.NotFound("Not found user", map[string]string{"user": "User not found"})
	ErrUserNotActive = errors.BadRequest("not active user", map[string]string{"active": "user not active yet"})

	ErrMissingToken   = errors.BadRequest("Token missing", map[string]string{"token": "Missing token"})
	ErrTokenGenerated = errors.InternalServerError("Token gen failed", "Generate token failed")
	ErrTokenInvalid   = errors.Unauthenticated("Invalid token", "token", "Token invalid")
	// ErrTokenNotFound  = errors.BadRequest("TOKEN_NOT_FOUND", "Token not found")
	// ErrTokenExpired   = errors.Unauthorized("TOKEN_EXPIRE", "Token expired")

	ErrConnectDB = errors.InternalServerError("Connect db failed", "Connecting to database failed")

	ErrMissingAccountID      = errors.BadRequest("Missing account id", map[string]string{"id": "Missing account id"})
	ErrInvalidAccountBalance = errors.BadRequest("Invalid account balance (>=0)", map[string]string{"balance": "greater than zero"})
	ErrAccountNotFound       = errors.NotFound("Not found user account", map[string]string{"account": "Account not found"})
	ErrUpdateAccountID       = errors.BadRequest("cannot update id", map[string]string{"update_mask": "cannot update id field"})
)
