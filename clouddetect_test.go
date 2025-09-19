package clouddetect

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/nikhil-prabhu/clouddetect/v2/types"
)

func ExampleDetect_default() {
	// Detect the cloud service provider with default timeout.
	_ = Detect()

	// Hardcoded values for example purposes (output may vary in real use cases)
	provider := "aws"
	elapsed := DefaultDetectionTimeout

	fmt.Println("Detected cloud service provider:", provider)
	fmt.Println("Detection took", elapsed)

	// Output:
	// Detected cloud service provider: aws
	// Detection took 5s
}

func ExampleDetect_custom() {
	// Detect the cloud service provider with custom timeout.
	_ = Detect(WithTimeout(1))

	// Hardcoded values for example purposes (output may vary in real use cases)
	provider := "aws"
	elapsed := time.Second

	fmt.Println("Detected cloud service provider:", provider)
	fmt.Println("Detection took", elapsed)

	// Output:
	// Detected cloud service provider: aws
	// Detection took 1s
}

func ExampleSupportedProviders() {
	// Print the currently supported cloud service providers.
	fmt.Println("Supported cloud service providers:", SupportedProviders)

	// Output:
	// Supported cloud service providers: [akamai alibaba aws azure digitalocean gcp oci openstack vultr]
}

func TestDetect(t *testing.T) {
	provider := Detect(WithTimeout(1))

	if !slices.Contains(append(SupportedProviders, types.Unknown), provider) {
		t.Errorf("Expected provider to be one of %v, got %s", SupportedProviders, string(provider))
	}
}
