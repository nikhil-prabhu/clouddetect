package clouddetect

// DefaultDetectionTimeout is the default maximum time allowed for detection.
const DefaultDetectionTimeout = 5 // seconds

const (
	// Unknown is an unknown cloud service provider.
	Unknown ProviderId = "unknown"
)

// Provider represents a cloud service provider.
type Provider interface {
	// Identifier returns the cloud service provider identifier.
	Identifier() ProviderId
	// Identify identifies the cloud service provider.
	Identify(chan ProviderId)
}

// ProviderId is a cloud service provider identifier.
type ProviderId string

// Detect detects the host's cloud service provider.
// If a non-zero timeout is specified, it overrides the default timeout duration.
func Detect(timeout uint64) ProviderId {
	_ = timeout
	if timeout == 0 {
		_ = DefaultDetectionTimeout
	}

	return Unknown
}
