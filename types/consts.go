package types

// ProviderId is a cloud service provider identifier.
type ProviderId string

const (
	// Unknown is an unknown cloud service provider.
	Unknown ProviderId = "unknown"
	// Alibaba is the Alibaba Cloud service provider.
	Alibaba = "alibaba"
	// Aws is the Amazon Web Services cloud service provider.
	Aws = "aws"
	// Azure is the Microsoft Azure cloud service provider.
	Azure = "azure"
)
