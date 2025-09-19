// Package clouddetect provides a simple way to detect the cloud service provider of a host.
package clouddetect

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/nikhil-prabhu/clouddetect/v2/providers/alibaba"
	"github.com/nikhil-prabhu/clouddetect/v2/providers/aws"
	"github.com/nikhil-prabhu/clouddetect/v2/providers/azure"
	"github.com/nikhil-prabhu/clouddetect/v2/providers/digitalocean"
	"github.com/nikhil-prabhu/clouddetect/v2/providers/gcp"
	"github.com/nikhil-prabhu/clouddetect/v2/providers/oci"
	"github.com/nikhil-prabhu/clouddetect/v2/providers/openstack"
	"github.com/nikhil-prabhu/clouddetect/v2/providers/vultr"
	"github.com/nikhil-prabhu/clouddetect/v2/types"
)

// DefaultDetectionTimeout is the default maximum time allowed for detection.
const DefaultDetectionTimeout = 5 * time.Second // seconds

// SupportedProviders is a list of supported cloud service providers.
var SupportedProviders = []types.ProviderId{
	types.Akamai,
	types.Alibaba,
	types.Aws,
	types.Azure,
	types.DigitalOcean,
	types.Gcp,
	types.Oci,
	types.OpenStack,
	types.Vultr,
}

type Option func(*config)

type config struct {
	timeout time.Duration
	logger  *zap.Logger
}

// Provider represents a cloud service provider.
//
// This interface is not guaranteed to remain stable/public and may change or be removed in the future.
// Do not depend on this interface outside of this package.
type Provider interface {
	Identifier() types.ProviderId                                   // Identifier returns the cloud service provider identifier.
	Identify(context.Context, chan<- types.ProviderId, *zap.Logger) // Identify detects the cloud service provider.
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

func WithTimeout(seconds uint64) Option {
	return func(c *config) {
		c.timeout = time.Duration(seconds) * time.Second
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(c *config) {
		c.logger = logger
	}
}

// Detect detects the host's cloud service provider.
// Options can be passed to customize the detection behavior, such as setting a custom timeout and logger.
func Detect(opts ...Option) types.ProviderId {
	// Default config
	cfg := config{
		timeout: DefaultDetectionTimeout,
		logger:  zap.NewNop(),
	}

	for _, o := range opts {
		o(&cfg)
	}

	ch := make(chan types.ProviderId, 1)
	wg := sync.WaitGroup{}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
	defer cancel()

	for name, provider := range providers {
		wg.Add(1)
		go func(name types.ProviderId, provider Provider) {
			cfg.logger.Debug(fmt.Sprintf("Starting detection routine for %s", name))
			defer wg.Done()
			provider.Identify(ctx, ch, cfg.logger)
		}(name, provider)
	}

	go func() {
		wg.Wait()
		cancel()
	}()

	select {
	case result := <-ch:
		cfg.logger.Info(fmt.Sprintf("Detected cloud service provider: %s", result))
		return result
	case <-ctx.Done():
		cfg.logger.Error(fmt.Sprintf("Detection timed out after %d seconds", cfg.timeout))
		return types.Unknown
	}
}
