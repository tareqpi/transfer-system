package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorObject struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	RequestID string      `json:"request_id"`
	Error     ErrorObject `json:"error"`
}

func WriteError(c *gin.Context, status int, code, message string) {
	requestID := c.GetString("request_id")
	if requestID == "" {
		requestID = c.GetHeader(headerRequestID)
	}
	resp := ErrorResponse{
		RequestID: requestID,
		Error:     ErrorObject{Code: code, Message: message},
	}
	c.AbortWithStatusJSON(status, resp)
}

func BadRequest(c *gin.Context, code, message string) {
	WriteError(c, http.StatusBadRequest, code, message)
}
func NotFound(c *gin.Context, code, message string) {
	WriteError(c, http.StatusNotFound, code, message)
}
func Conflict(c *gin.Context, code, message string) {
	WriteError(c, http.StatusConflict, code, message)
}
func Internal(c *gin.Context, message string) {
	WriteError(c, http.StatusInternalServerError, "internal_error", message)
}
