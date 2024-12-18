package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSetLogger(t *testing.T) {
	defaultLogger := Logger
	assert.IsType(t, &zap.Logger{}, defaultLogger, "Default logger should be of type *zap.Logger")
	assert.Equal(t, zap.NewNop(), defaultLogger, "Default logger should be a no-op logger")

	newLogger, err := zap.NewProduction()
	if err != nil {
		t.Fatalf("Failed to create a new logger: %v", err)
	}

	SetLogger(newLogger)

	assert.Equal(t, newLogger, Logger, "Logger should be updated to the new logger")

	SetLogger(zap.NewNop())
}
