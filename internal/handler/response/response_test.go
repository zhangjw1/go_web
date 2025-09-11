package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("request_id", "test-request-id")
	return c, w
}

func TestSuccess(t *testing.T) {
	c, w := setupTestContext()
	testData := map[string]string{"key": "value"}

	Success(c, testData, "Test success message")

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, "Test success message", response.Message)
	assert.Equal(t, "test-request-id", response.RequestID)
	assert.NotEmpty(t, response.Timestamp)
	assert.NotNil(t, response.Data)
}

func TestSuccessWithMeta(t *testing.T) {
	c, w := setupTestContext()
	testData := []string{"item1", "item2"}
	meta := &Meta{
		Page:       1,
		PerPage:    10,
		Total:      2,
		TotalPages: 1,
	}

	SuccessWithMeta(c, testData, meta)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, "Success", response.Message)
	assert.NotNil(t, response.Meta)
	assert.Equal(t, 1, response.Meta.Page)
	assert.Equal(t, 10, response.Meta.PerPage)
	assert.Equal(t, 2, response.Meta.Total)
	assert.Equal(t, 1, response.Meta.TotalPages)
}

func TestCreated(t *testing.T) {
	c, w := setupTestContext()
	testData := map[string]interface{}{"id": 1, "name": "test"}

	Created(c, testData)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, "Created successfully", response.Message)
	assert.NotNil(t, response.Data)
}

func TestNoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	// Create a proper request
	req, _ := http.NewRequest("DELETE", "/test", nil)
	c.Request = req
	c.Set("request_id", "test-request-id")

	NoContent(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestBadRequest(t *testing.T) {
	c, w := setupTestContext()
	details := map[string]string{"field": "username", "issue": "required"}

	BadRequest(c, "Validation failed", details)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "BAD_REQUEST", response.Error.Code)
	assert.Equal(t, "Validation failed", response.Error.Message)
	assert.NotNil(t, response.Error.Details)
}

func TestUnauthorized(t *testing.T) {
	c, w := setupTestContext()

	Unauthorized(c, "Invalid credentials")

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "UNAUTHORIZED", response.Error.Code)
	assert.Equal(t, "Invalid credentials", response.Error.Message)
}

func TestForbidden(t *testing.T) {
	c, w := setupTestContext()

	Forbidden(c)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "FORBIDDEN", response.Error.Code)
	assert.Equal(t, "Forbidden", response.Error.Message)
}

func TestNotFound(t *testing.T) {
	c, w := setupTestContext()

	NotFound(c, "User not found")

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "NOT_FOUND", response.Error.Code)
	assert.Equal(t, "User not found", response.Error.Message)
}

func TestConflict(t *testing.T) {
	c, w := setupTestContext()
	details := map[string]string{"field": "email", "value": "test@example.com"}

	Conflict(c, "Email already exists", details)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "CONFLICT", response.Error.Code)
	assert.Equal(t, "Email already exists", response.Error.Message)
	assert.NotNil(t, response.Error.Details)
}

func TestUnprocessableEntity(t *testing.T) {
	c, w := setupTestContext()
	validationErrors := []map[string]string{
		{"field": "email", "message": "invalid format"},
		{"field": "password", "message": "too short"},
	}

	UnprocessableEntity(c, "Validation errors", validationErrors)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "UNPROCESSABLE_ENTITY", response.Error.Code)
	assert.Equal(t, "Validation errors", response.Error.Message)
	assert.NotNil(t, response.Error.Details)
}

func TestInternalServerError(t *testing.T) {
	c, w := setupTestContext()

	InternalServerError(c, "Database connection failed")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Error.Code)
	assert.Equal(t, "Database connection failed", response.Error.Message)
}

func TestServiceUnavailable(t *testing.T) {
	c, w := setupTestContext()

	ServiceUnavailable(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "SERVICE_UNAVAILABLE", response.Error.Code)
	assert.Equal(t, "Service unavailable", response.Error.Message)
}

func TestError(t *testing.T) {
	c, w := setupTestContext()
	details := map[string]interface{}{"retry_after": 30}

	Error(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests", details)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "RATE_LIMIT_EXCEEDED", response.Error.Code)
	assert.Equal(t, "Too many requests", response.Error.Message)
	assert.NotNil(t, response.Error.Details)
}

func TestResponseStructure(t *testing.T) {
	c, w := setupTestContext()

	Success(c, nil)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Test that all expected fields are present
	assert.True(t, response.Success)
	assert.Equal(t, "Success", response.Message)
	assert.Equal(t, "test-request-id", response.RequestID)
	assert.NotEmpty(t, response.Timestamp)
	assert.Nil(t, response.Error)
	assert.Nil(t, response.Meta)

	// Test timestamp format (should be RFC3339)
	_, err = json.Marshal(response.Timestamp)
	assert.NoError(t, err)
}

func BenchmarkSuccess(b *testing.B) {
	c, _ := setupTestContext()
	testData := map[string]string{"key": "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Success(c, testData)
	}
}

func BenchmarkBadRequest(b *testing.B) {
	c, _ := setupTestContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BadRequest(c, "Test error message")
	}
}