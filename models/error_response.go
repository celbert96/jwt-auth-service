package models

import (
	"fmt"
	"net/http"
)

type ErrorResponse struct {
	ErrorMessage string `json:"error_message"`
}

func ErrResponseForHttpStatus(status int) ErrorResponse {
	switch status {
	case http.StatusForbidden:
		return ErrorResponse{ErrorMessage: "access denied"}
	case http.StatusNotFound:
		return ErrorResponse{ErrorMessage: "resource not found"}
	case http.StatusBadRequest:
		return ErrorResponse{ErrorMessage: "bad request"}
	case http.StatusInternalServerError:
		return ErrorResponse{ErrorMessage: "internal server error"}
	default:
		return ErrorResponse{ErrorMessage: fmt.Sprintf("unknown error: code %d", status)}
	}
}
