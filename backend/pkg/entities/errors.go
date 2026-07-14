package entities

type ErrorCode string

const (
	ErrValidation       ErrorCode = "VALIDATION_ERR"
	ErrNotAuthenticated ErrorCode = "NOT_AUTHENTICATED"
	ErrForbidded        ErrorCode = "FORBIDDEN"
	ErrNotFound         ErrorCode = "NOT_FOUND"
	ErrInternal         ErrorCode = "INTERNAL_ERR"
)

type Error struct {
	Code    ErrorCode      `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"details,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}

func NewError(code ErrorCode, message string, data ...map[string]any) *Error {
	errEntity := &Error{
		Code:    code,
		Message: message,
	}
	if len(data) > 0 {
		errEntity.Data = data[0]
	}
	return errEntity
}

func IsDomainError(err error) (*Error, bool) {
	e, ok := err.(*Error)
	return e, ok
}
