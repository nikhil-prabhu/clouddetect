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
	// DigitalOcean is the DigitalOcean cloud service provider.
	DigitalOcean = "digitalocean"
	// Gcp is the Google Cloud Platform cloud service provider.
	Gcp = "gcp"
	// Oci is the Oracle Cloud Infrastructure cloud service provider.
	Oci = "oci"
	// OpenStack is the OpenStack cloud service provider.
	OpenStack = "openstack"
	// Vultr is the Vultr cloud service provider.
	Vultr = "vultr"
)
