package api

type ResponseError struct {
	Code    int
	Message string
}

func (r ResponseError) Error() string { return r.Message }

func NewResponseError(code int, msg string) *ResponseError {
	return &ResponseError{
		Code:    code,
		Message: msg,
	}
}
