package types

// ProviderId is a cloud service provider identifier.
type ProviderId string

const (
	Unknown      ProviderId = "unknown"      // Unknown is the unknown cloud service provider.
	Alibaba      ProviderId = "alibaba"      // Alibaba is the Alibaba Cloud service provider.
	Aws          ProviderId = "aws"          // Aws is the Amazon Web Services cloud service provider.
	Azure        ProviderId = "azure"        // Azure is the Microsoft Azure cloud service provider.
	DigitalOcean ProviderId = "digitalocean" // DigitalOcean is the DigitalOcean cloud service provider.
	Gcp          ProviderId = "gcp"          // Gcp is the Google Cloud Platform cloud service provider.
	Oci          ProviderId = "oci"          // Oci is the Oracle Cloud Infrastructure cloud service provider.
	OpenStack    ProviderId = "openstack"    // OpenStack is the OpenStack cloud service provider.
	Vultr        ProviderId = "vultr"        // Vultr is the Vultr cloud service provider.
)
