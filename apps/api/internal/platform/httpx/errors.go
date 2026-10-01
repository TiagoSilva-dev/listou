package httpx

import (
	"errors"
	"fmt"
	"net/http"
)

// Error is an application error that maps to a stable API error code.
// Domain packages return these (or wrap them); handlers never build
// status codes by hand.
type Error struct {
	Status  int
	Code    string
	Message string
	Fields  map[string]string
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return e.Code + ": " + e.Message
}

func (e *Error) Unwrap() error { return e.cause }

func (e *Error) WithCause(err error) *Error {
	cp := *e
	cp.cause = err
	return &cp
}

func NewError(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}

func NotFound(code, message string) *Error {
	return NewError(http.StatusNotFound, code, message)
}

func BadRequest(code, message string) *Error {
	return NewError(http.StatusBadRequest, code, message)
}

func Validation(fields map[string]string) *Error {
	return &Error{Status: http.StatusUnprocessableEntity, Code: "VALIDATION_FAILED", Message: "Alguns campos precisam de atenção.", Fields: fields}
}

var (
	ErrUnauthorized = NewError(http.StatusUnauthorized, "UNAUTHORIZED", "Você precisa entrar para continuar.")
	ErrForbidden    = NewError(http.StatusForbidden, "FORBIDDEN", "Você não tem permissão para esta ação.")
	ErrInternal     = NewError(http.StatusInternalServerError, "INTERNAL_ERROR", "Algo deu errado. Tente novamente.")
	ErrRateLimited  = NewError(http.StatusTooManyRequests, "RATE_LIMITED", "Muitas tentativas. Aguarde um pouco.")
)

// AsError extracts an *Error from err, defaulting to ErrInternal.
func AsError(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return ErrInternal
}
