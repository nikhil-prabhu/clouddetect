package clouddetect

import (
	"fmt"
	"slices"
	"testing"

	"github.com/nikhil-prabhu/clouddetect/types"
)

func ExampleDetect_default() {
	// Detect the cloud service provider with default timeout.
	_ = Detect(0)

	// Hardcoded values for example purposes (output may vary in real use cases)
	provider := "aws"
	elapsed := DefaultDetectionTimeout

	fmt.Println("Detected cloud service provider:", provider)
	fmt.Println("Detection took", elapsed, "seconds")

	// Output:
	// Detected cloud service provider: aws
	// Detection took 5 seconds
}

func ExampleDetect_custom() {
	// Detect the cloud service provider with custom timeout.
	_ = Detect(1)

	// Hardcoded values for example purposes (output may vary in real use cases)
	provider := "aws"
	elapsed := 1

	fmt.Println("Detected cloud service provider:", provider)
	fmt.Println("Detection took", elapsed, "second")

	// Output:
	// Detected cloud service provider: aws
	// Detection took 1 second
}

func ExampleSupportedProviders() {
	// Print the currently supported cloud service providers.
	fmt.Println("Supported cloud service providers:", SupportedProviders)

	// Output:
	// Supported cloud service providers: [akamai alibaba aws azure digitalocean gcp oci openstack vultr]
}

func TestDetect(t *testing.T) {
	provider := Detect(1)

	if !slices.Contains(append(SupportedProviders, types.Unknown), provider) {
		t.Errorf("Expected provider to be one of %v, got %s", SupportedProviders, string(provider))
	}
}
