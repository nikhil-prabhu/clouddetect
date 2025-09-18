// Package digitalocean implements the DigitalOcean cloud provider detection.
package digitalocean

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/nikhil-prabhu/clouddetect/logging"
	"github.com/nikhil-prabhu/clouddetect/types"
)

const (
	metadataURL string = "http://169.254.169.254/metadata/v1.json"
	vendorFile         = "/sys/class/dmi/id/sys_vendor"
	identifier         = types.DigitalOcean
)

type metadataResponse struct {
	DropletID uint `json:"droplet_id"`
}

type DigitalOcean struct{}

func (d *DigitalOcean) Identifier() types.ProviderId {
	return identifier
}

func (d *DigitalOcean) Identify(ctx context.Context, ch chan<- types.ProviderId) {
	if d.checkMetadataServer(ctx) {
		ch <- d.Identifier()
		return
	}

	if d.checkVendorFile(vendorFile) {
		ch <- d.Identifier()
		return
	}
}

func (d *DigitalOcean) checkMetadataServer(ctx context.Context) bool {
	logging.Logger.Debug(fmt.Sprintf("Checking %s metadata using url %s", identifier, metadataURL))

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", metadataURL, nil)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("Error creating request: %s", err))
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("Error reading response: %s", err))
		return false
	}
	defer func(Body io.ReadCloser) {
		closeErr := Body.Close()
		if closeErr != nil {
			logging.Logger.Error(fmt.Sprintf("Error closing response body: %s", closeErr))
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		logging.Logger.Error(fmt.Sprintf("Error response status code: %d", resp.StatusCode))
		return false
	}

	metadata := new(metadataResponse)
	if decodeErr := json.NewDecoder(resp.Body).Decode(metadata); decodeErr != nil {
		logging.Logger.Error(fmt.Sprintf("Error decoding response: %s", decodeErr))
		return false
	}

	return metadata.DropletID > 0
}

func (d *DigitalOcean) checkVendorFile(file string) bool {
	logging.Logger.Debug(fmt.Sprintf("Checking %s vendor file %s", identifier, file))

	content, err := os.ReadFile(file)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("Error reading file: %s", err))
		return false
	}

	return strings.Contains(string(content), "DigitalOcean")
}
