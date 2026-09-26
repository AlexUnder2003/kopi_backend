package apperrors

type Kind int

const (
	KindNotFound Kind = iota + 1
	KindConflict
	KindBadRequest
	KindInternalServerError
	KindUnauthorized
	KindForbidden
)

type AppError struct {
	Kind    Kind
	Message string
}

func (e *AppError) Error() string { return e.Message }

func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Kind == t.Kind
}

func New(kind Kind, message string) *AppError {
	return &AppError{Kind: kind, Message: message}
}

func NotFound(message string) *AppError {
	return New(KindNotFound, message)
}

func Internal(message string) *AppError {
	return New(KindInternalServerError, message)
}

func Unauthorized(message string) *AppError {
	return New(KindUnauthorized, message)
}

func Forbidden(message string) *AppError {
	return New(KindForbidden, message)
}

func BadRequest(message string) *AppError {
	return New(KindBadRequest, message)
}

func Conflict(message string) *AppError {
	return New(KindConflict, message)
}

var (
	ErrNotFound            = &AppError{Kind: KindNotFound}
	ErrConflict            = &AppError{Kind: KindConflict}
	ErrBadRequest          = &AppError{Kind: KindBadRequest}
	ErrInternalServerError = &AppError{Kind: KindInternalServerError}
	ErrUnauthorized        = &AppError{Kind: KindUnauthorized}
	ErrForbidden           = &AppError{Kind: KindForbidden}
)
