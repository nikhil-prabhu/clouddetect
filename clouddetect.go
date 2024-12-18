package clouddetect

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nikhil-prabhu/clouddetect/logging"
	"github.com/nikhil-prabhu/clouddetect/providers/alibaba"
	"github.com/nikhil-prabhu/clouddetect/providers/aws"
	"github.com/nikhil-prabhu/clouddetect/providers/azure"
	"github.com/nikhil-prabhu/clouddetect/providers/digitalocean"
	"github.com/nikhil-prabhu/clouddetect/providers/gcp"
	"github.com/nikhil-prabhu/clouddetect/providers/oci"
	"github.com/nikhil-prabhu/clouddetect/providers/openstack"
	"github.com/nikhil-prabhu/clouddetect/providers/vultr"
	"github.com/nikhil-prabhu/clouddetect/types"
)

// DefaultDetectionTimeout is the default maximum time allowed for detection.
const DefaultDetectionTimeout = 5 // seconds

// SupportedProviders is a list of supported cloud service providers.
var SupportedProviders = []types.ProviderId{
	types.Alibaba,
	types.Aws,
	types.Azure,
	types.DigitalOcean,
	types.Gcp,
	types.Oci,
	types.OpenStack,
	types.Vultr,
}

// Provider represents a cloud service provider.
type Provider interface {
	// Identifier returns the cloud service provider identifier.
	Identifier() types.ProviderId
	// Identify identifies the cloud service provider.
	Identify(chan<- types.ProviderId)
}

var providers = map[types.ProviderId]Provider{
	types.Alibaba:      &alibaba.Alibaba{},
	types.Aws:          &aws.Aws{},
	types.Azure:        &azure.Azure{},
	types.DigitalOcean: &digitalocean.DigitalOcean{},
	types.Gcp:          &gcp.Gcp{},
	types.Oci:          &oci.Oci{},
	types.OpenStack:    &openstack.OpenStack{},
	types.Vultr:        &vultr.Vultr{},
}

// Detect detects the host's cloud service provider.
// If a non-zero timeout is specified, it overrides the default timeout duration.
func Detect(timeout uint64) types.ProviderId {
	t := timeout
	if timeout == 0 {
		t = DefaultDetectionTimeout
	}

	ch := make(chan types.ProviderId, 1)
	wg := sync.WaitGroup{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(t)*time.Second)
	defer cancel()

	for name, provider := range providers {
		wg.Add(1)
		go func(name types.ProviderId, provider Provider) {
			logging.Logger.Debug(fmt.Sprintf("Starting detection routine for %s", name))
			defer wg.Done()
			provider.Identify(ch)
		}(name, provider)
	}

	go func() {
		wg.Wait()
		cancel()
	}()

	select {
	case result := <-ch:
		logging.Logger.Info(fmt.Sprintf("Detected cloud service provider: %s", result))
		return result
	case <-ctx.Done():
		logging.Logger.Error(fmt.Sprintf("Detection timed out after %d seconds", t))
		return types.Unknown
	}
}
