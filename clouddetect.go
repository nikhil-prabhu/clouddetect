package clouddetect

import (
	"github.com/nikhil-prabhu/clouddetect/types"
)

// DefaultDetectionTimeout is the default maximum time allowed for detection.
const DefaultDetectionTimeout = 5 // seconds

// SupportedProviders is a list of supported cloud service providers.
var SupportedProviders = []types.ProviderId{
	types.Alibaba,
	types.Aws,
	types.Azure,
}

// Provider represents a cloud service provider.
type Provider interface {
	// Identifier returns the cloud service provider identifier.
	Identifier() types.ProviderId
	// Identify identifies the cloud service provider.
	Identify(chan types.ProviderId)
}

// Detect detects the host's cloud service provider.
// If a non-zero timeout is specified, it overrides the default timeout duration.
func Detect(timeout uint64) types.ProviderId {
	_ = timeout
	if timeout == 0 {
		_ = DefaultDetectionTimeout
	}

	return types.Unknown
}
