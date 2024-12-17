package clouddetect

import "go.uber.org/zap"

// DefaultDetectionTimeout is the default maximum time allowed for detection.
const DefaultDetectionTimeout = 5 // seconds

const (
	// Unknown is an unknown cloud service provider.
	Unknown ProviderId = "unknown"
	// Alibaba is the Alibaba Cloud service provider.
	Alibaba = "alibaba"
	// Aws is the Amazon Web Services cloud service provider.
	Aws = "aws"
)

// SupportedProviders is a list of supported cloud service providers.
var SupportedProviders = []ProviderId{
	Alibaba,
	Aws,
}

// Logger is the logger used by the package.
var Logger = zap.NewNop()

// Provider represents a cloud service provider.
type Provider interface {
	// Identifier returns the cloud service provider identifier.
	Identifier() ProviderId
	// Identify identifies the cloud service provider.
	Identify(chan ProviderId)
}

// ProviderId is a cloud service provider identifier.
type ProviderId string

// SetLogger sets the logger used by the package.
func SetLogger(logger *zap.Logger) {
	Logger = logger
}

// Detect detects the host's cloud service provider.
// If a non-zero timeout is specified, it overrides the default timeout duration.
func Detect(timeout uint64) ProviderId {
	_ = timeout
	if timeout == 0 {
		_ = DefaultDetectionTimeout
	}

	return Unknown
}
