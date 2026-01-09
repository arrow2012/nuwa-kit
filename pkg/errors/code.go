package errors

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorCode is the interface for business errors
type ErrorCode interface {
	HTTPStatus() int
	BusinessCode() int
	Message() string
	Error() string // To satisfy standard error interface
}

type customError struct {
	httpStatus   int
	businessCode int
	message      string
}

func (e *customError) HTTPStatus() int {
	return e.httpStatus
}

func (e *customError) BusinessCode() int {
	return e.businessCode
}

func (e *customError) Message() string {
	return e.message
}

func (e *customError) Error() string {
	return e.message
}

// New creates a new custom error
func New(httpStatus, businessCode int, message string) ErrorCode {
	return &customError{
		httpStatus:   httpStatus,
		businessCode: businessCode,
		message:      message,
	}
}

// GRPCStatus implements the GRPCStatus interface found in google.golang.org/grpc/status
func (e *customError) GRPCStatus() *status.Status {
	var c codes.Code
	switch e.httpStatus {
	case http.StatusOK:
		c = codes.OK
	case http.StatusBadRequest:
		c = codes.InvalidArgument
	case http.StatusUnauthorized:
		c = codes.Unauthenticated
	case http.StatusForbidden:
		c = codes.PermissionDenied
	case http.StatusNotFound:
		c = codes.NotFound
	case http.StatusConflict:
		c = codes.AlreadyExists
	case http.StatusTooManyRequests:
		c = codes.ResourceExhausted
	case http.StatusInternalServerError:
		c = codes.Internal
	case http.StatusServiceUnavailable:
		c = codes.Unavailable
	case http.StatusGatewayTimeout:
		c = codes.DeadlineExceeded
	default:
		c = codes.Unknown
	}
	return status.New(c, e.message)
}

// Common Errors
var (
	Success             = New(http.StatusOK, 0, "success")
	ErrInternalServer   = New(http.StatusInternalServerError, 10001, "internal server error")
	ErrInvalidParams    = New(http.StatusBadRequest, 10002, "invalid parameters")
	ErrUnauthorized     = New(http.StatusUnauthorized, 10003, "unauthorized")
	ErrNotFound         = New(http.StatusNotFound, 10004, "resource not found")
	ErrMethodNotAllowed = New(http.StatusMethodNotAllowed, 10005, "method not allowed")
	ErrTooManyRequests  = New(http.StatusTooManyRequests, 10006, "too many requests")

	// Auth Errors
	ErrUserNotFound       = New(http.StatusNotFound, 20001, "user not found")
	ErrPasswordIncorrect  = New(http.StatusUnauthorized, 20002, "password incorrect")
	ErrTokenInvalid       = New(http.StatusUnauthorized, 20003, "token invalid")
	ErrTokenExpired       = New(http.StatusUnauthorized, 20004, "token expired")
	ErrUserAlreadyExists  = New(409, 20409, "User already exists")
	ErrInvalidCredentials = New(401, 20401, "Invalid credentials")
	ErrMFARequired        = New(403, 20403, "MFA required")
	ErrPasswordExpired    = New(403, 20404, "Password expired")
	ErrForbidden          = New(http.StatusForbidden, 10007, "forbidden")
)
