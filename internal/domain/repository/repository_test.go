package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepositoryError(t *testing.T) {
	t.Run("NewRepositoryError", func(t *testing.T) {
		err := NewRepositoryError("create", "User", 123, "test error", nil)
		assert.NotNil(t, err)
		assert.Equal(t, "create", err.Op)
		assert.Equal(t, "User", err.Entity)
		assert.Equal(t, uint(123), err.ID)
		assert.Equal(t, "test error", err.Message)
	})

	t.Run("Error message with ID", func(t *testing.T) {
		err := NewRepositoryError("get", "User", 123, "not found", nil)
		errMsg := err.Error()
		assert.Contains(t, errMsg, "get")
		assert.Contains(t, errMsg, "User")
		assert.Contains(t, errMsg, "not found")
	})

	t.Run("Error message without ID", func(t *testing.T) {
		err := NewRepositoryError("create", "User", 0, "validation failed", nil)
		errMsg := err.Error()
		assert.Contains(t, errMsg, "create")
		assert.Contains(t, errMsg, "User")
		assert.Contains(t, errMsg, "validation failed")
	})

	t.Run("IsNotFoundError", func(t *testing.T) {
		notFoundErr := NewRepositoryError("get", "User", 123, "entity not found", nil)
		otherErr := NewRepositoryError("create", "User", 0, "validation failed", nil)

		assert.True(t, IsNotFoundError(notFoundErr))
		assert.False(t, IsNotFoundError(otherErr))
	})

	t.Run("IsAlreadyExistsError", func(t *testing.T) {
		existsErr := NewRepositoryError("create", "User", 0, "entity already exists", nil)
		otherErr := NewRepositoryError("get", "User", 123, "not found", nil)

		assert.True(t, IsAlreadyExistsError(existsErr))
		assert.False(t, IsAlreadyExistsError(otherErr))
	})
}

func TestRepositoryErrorUnwrap(t *testing.T) {
	originalErr := assert.AnError
	repoErr := NewRepositoryError("test", "Entity", 1, "test message", originalErr)
	
	assert.Equal(t, originalErr, repoErr.Unwrap())
}