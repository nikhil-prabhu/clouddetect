package azure

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
	metadataURL string = "http://169.254.169.254/metadata/instance?api-version=2017-12-01"
	vendorFile         = "/sys/class/dmi/id/sys_vendor"
	identifier         = types.Azure
)

type compute struct {
	VmID string `json:"vmId"`
}

type metadataResponse struct {
	Compute compute `json:"compute"`
}

type Azure struct{}

func (a *Azure) Identifier() types.ProviderId {
	return identifier
}

func (a *Azure) Identify(ctx context.Context, ch chan<- types.ProviderId) {
	if a.checkMetadataServer(ctx) {
		ch <- a.Identifier()
		return
	}

	if a.checkVendorFile(vendorFile) {
		ch <- a.Identifier()
		return
	}
}

func (a *Azure) checkMetadataServer(ctx context.Context) bool {
	logging.Logger.Debug(fmt.Sprintf("Checking %s metadata using url %s", identifier, metadataURL))

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", metadataURL, nil)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("Error creating request: %s", err))
		return false
	}
	req.Header.Add("Metadata", "true")

	resp, err := client.Do(req)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("Error sending request: %s", err))
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

	return len(metadata.Compute.VmID) > 0
}

func (a *Azure) checkVendorFile(file string) bool {
	logging.Logger.Debug(fmt.Sprintf("Checking %s vendor file %s", identifier, file))

	content, err := os.ReadFile(file)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("Error reading file: %s", err))
		return false
	}

	return strings.Contains(string(content), "Microsoft Corporation")
}
