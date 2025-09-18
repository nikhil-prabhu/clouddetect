// Package logging provides a simple logging utility using Uber's zap library.
package logging

import "go.uber.org/zap"

// Logger is the logger used by the package.
var Logger = zap.NewNop()
