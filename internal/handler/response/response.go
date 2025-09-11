package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response represents the standard API response format
type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// ErrorInfo represents error information in the response
type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Meta represents metadata in the response
type Meta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// Success sends a successful response
func Success(c *gin.Context, data interface{}, message ...string) {
	msg := "Success"
	if len(message) > 0 {
		msg = message[0]
	}

	response := Response{
		Success:   true,
		Message:   msg,
		Data:      data,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// SuccessWithMeta sends a successful response with metadata
func SuccessWithMeta(c *gin.Context, data interface{}, meta *Meta, message ...string) {
	msg := "Success"
	if len(message) > 0 {
		msg = message[0]
	}

	response := Response{
		Success:   true,
		Message:   msg,
		Data:      data,
		Meta:      meta,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.AbortWithStatus(http.StatusNoContent)
}

// Created sends a 201 Created response
func Created(c *gin.Context, data interface{}, message ...string) {
	msg := "Created successfully"
	if len(message) > 0 {
		msg = message[0]
	}

	response := Response{
		Success:   true,
		Message:   msg,
		Data:      data,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusCreated, response)
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string, details ...interface{}) {
	errorInfo := &ErrorInfo{
		Code:    "BAD_REQUEST",
		Message: message,
	}

	if len(details) > 0 {
		errorInfo.Details = details[0]
	}

	response := Response{
		Success:   false,
		Error:     errorInfo,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusBadRequest, response)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message ...string) {
	msg := "Unauthorized"
	if len(message) > 0 {
		msg = message[0]
	}

	errorInfo := &ErrorInfo{
		Code:    "UNAUTHORIZED",
		Message: msg,
	}

	response := Response{
		Success:   false,
		Error:     errorInfo,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusUnauthorized, response)
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message ...string) {
	msg := "Forbidden"
	if len(message) > 0 {
		msg = message[0]
	}

	errorInfo := &ErrorInfo{
		Code:    "FORBIDDEN",
		Message: msg,
	}

	response := Response{
		Success:   false,
		Error:     errorInfo,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusForbidden, response)
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, message ...string) {
	msg := "Resource not found"
	if len(message) > 0 {
		msg = message[0]
	}

	errorInfo := &ErrorInfo{
		Code:    "NOT_FOUND",
		Message: msg,
	}

	response := Response{
		Success:   false,
		Error:     errorInfo,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusNotFound, response)
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, message string, details ...interface{}) {
	errorInfo := &ErrorInfo{
		Code:    "CONFLICT",
		Message: message,
	}

	if len(details) > 0 {
		errorInfo.Details = details[0]
	}

	response := Response{
		Success:   false,
		Error:     errorInfo,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusConflict, response)
}

// UnprocessableEntity sends a 422 Unprocessable Entity response
func UnprocessableEntity(c *gin.Context, message string, details ...interface{}) {
	errorInfo := &ErrorInfo{
		Code:    "UNPROCESSABLE_ENTITY",
		Message: message,
	}

	if len(details) > 0 {
		errorInfo.Details = details[0]
	}

	response := Response{
		Success:   false,
		Error:     errorInfo,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusUnprocessableEntity, response)
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *gin.Context, message ...string) {
	msg := "Internal server error"
	if len(message) > 0 {
		msg = message[0]
	}

	errorInfo := &ErrorInfo{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: msg,
	}

	response := Response{
		Success:   false,
		Error:     errorInfo,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusInternalServerError, response)
}

// ServiceUnavailable sends a 503 Service Unavailable response
func ServiceUnavailable(c *gin.Context, message ...string) {
	msg := "Service unavailable"
	if len(message) > 0 {
		msg = message[0]
	}

	errorInfo := &ErrorInfo{
		Code:    "SERVICE_UNAVAILABLE",
		Message: msg,
	}

	response := Response{
		Success:   false,
		Error:     errorInfo,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusServiceUnavailable, response)
}

// Error sends an error response with custom status code
func Error(c *gin.Context, statusCode int, code, message string, details ...interface{}) {
	errorInfo := &ErrorInfo{
		Code:    code,
		Message: message,
	}

	if len(details) > 0 {
		errorInfo.Details = details[0]
	}

	response := Response{
		Success:   false,
		Error:     errorInfo,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(statusCode, response)
}