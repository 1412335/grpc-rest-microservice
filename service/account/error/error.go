package error

import "github.com/1412335/grpc-rest-microservice/pkg/errors"

var (
	ErrMissingUsername = errors.BadRequest("MISSING_USERNAME", map[string]string{"username": "Missing username"})
	ErrMissingFullname = errors.BadRequest("MISSING_FULLNAME", map[string]string{"fullname": "Missing fullname"})

	ErrMissingEmail   = errors.BadRequest("Email is required", map[string]string{"email": "Missing email"})
	ErrInvalidEmail   = errors.BadRequest("Invalid email", map[string]string{"email": "The email provided is invalid"})
	ErrDuplicateEmail = errors.BadRequest("Duplicate email", map[string]string{"email": "A user with this email address already exists"})

	ErrInvalidPassword   = errors.BadRequest("Invalid password", map[string]string{"password": "Password must be at least 8 characters long"})
	ErrIncorrectPassword = errors.Unauthenticated("Email or password is incorrect", "password", "Email or password is incorrect")
	ErrHashPassword      = errors.InternalServerError("Hash password failed", "hash password failed")

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
)
