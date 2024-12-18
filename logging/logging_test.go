package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSetLogger(t *testing.T) {
	// Step 1: Verify the default logger is a no-op logger
	defaultLogger := Logger
	assert.IsType(t, &zap.Logger{}, defaultLogger, "Default logger should be of type *zap.Logger")
	assert.Equal(t, zap.NewNop(), defaultLogger, "Default logger should be a no-op logger")

	// Step 2: Create a new logger using zap.NewProduction() or any logger of your choice
	newLogger, err := zap.NewProduction()
	if err != nil {
		t.Fatalf("Failed to create a new logger: %v", err)
	}

	// Step 3: Call SetLogger to update the global Logger
	SetLogger(newLogger)

	// Step 4: Assert that the global Logger has been updated to the new logger
	assert.Equal(t, newLogger, Logger, "Logger should be updated to the new logger")

	// Step 5: Clean up by resetting to a no-op logger (optional)
	SetLogger(zap.NewNop())
}
