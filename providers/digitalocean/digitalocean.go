package digitalocean

import (
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
	return types.DigitalOcean
}

func (d *DigitalOcean) Identify(ch chan<- types.ProviderId) {
	if d.checkMetadataServer() {
		ch <- d.Identifier()
		return
	}

	if d.checkVendorFile(vendorFile) {
		ch <- d.Identifier()
		return
	}
}

func (d *DigitalOcean) checkMetadataServer() bool {
	logging.Logger.Debug(fmt.Sprintf("Checking %s metadata using url %s", identifier, metadataURL))

	resp, err := http.Get(metadataURL)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("Error reading response: %s", err))
		return false
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logging.Logger.Error(fmt.Sprintf("Error closing response body: %s", err))
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		logging.Logger.Error(fmt.Sprintf("Error response status code: %d", resp.StatusCode))
		return false
	}

	metadata := new(metadataResponse)
	if err = json.NewDecoder(resp.Body).Decode(metadata); err != nil {
		logging.Logger.Error(fmt.Sprintf("Error decoding response: %s", err))
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
