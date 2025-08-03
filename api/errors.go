package api

var (
	ErrTokenInvalid = CreateDefaultMessage("invalid token", 1000)
	ErrTokenExpired = CreateDefaultMessage("token expired", 1001)
	ErrTokenMissing = CreateDefaultMessage("token missing", 1002)
	ErrRequestBodyInvalid = CreateDefaultMessage("invalid request body", 1003)
	ErrSomethingWentWrongServer = CreateDefaultMessage("something went wrong", 1004)
)

type DefaultMessage struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type ResponseError struct {
	Error  string `json:"error"`
	Code   int    `json:"code"`
	Status int    `json:"status"`
}

func CreateDefaultMessage(message string, code int) DefaultMessage {
	return DefaultMessage{
		Message: message,
		Code:    code,
	}
}
func CreateErrorResponse(message string, status int, code int) ResponseError {
	return ResponseError{
		Error:  message,
		Code:   code,
		Status: status,
	}
}