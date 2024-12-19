package vultr

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
	metadataURL string = "http://169.254.169.254/v1.json"
	vendorFile         = "/sys/class/dmi/id/sys_vendor"
	identifier         = types.Vultr
)

type metadataResponse struct {
	InstanceID string `json:"instanceid"`
}

type Vultr struct{}

func (v *Vultr) Identifier() types.ProviderId {
	return identifier
}

func (v *Vultr) Identify(ctx context.Context, ch chan<- types.ProviderId) {
	if v.checkMetadataServer(ctx) {
		ch <- v.Identifier()
		return
	}

	if v.checkVendorFile(vendorFile) {
		ch <- v.Identifier()
		return
	}
}

func (v *Vultr) checkMetadataServer(ctx context.Context) bool {
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

	return len(metadata.InstanceID) > 0
}

func (v *Vultr) checkVendorFile(file string) bool {
	logging.Logger.Debug(fmt.Sprintf("Checking %s vendor file %s", identifier, file))

	content, err := os.ReadFile(file)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("Error reading file: %s", err))
		return false
	}

	return strings.Contains(string(content), "Vultr")
}
