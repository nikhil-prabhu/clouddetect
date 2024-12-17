package logging

import "go.uber.org/zap"

// Logger is the logger used by the package.
var Logger = zap.NewNop()

// SetLogger sets the logger used by the package.
func SetLogger(logger *zap.Logger) {
	Logger = logger
}
