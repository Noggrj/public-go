package infrastructure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsoleEmailService_SendEmail(t *testing.T) {
	service := NewConsoleEmailService()
	err := service.SendEmail("test@example.com", "Test Subject", "Test Body")
	assert.NoError(t, err)
	assert.NotNil(t, service)
}
