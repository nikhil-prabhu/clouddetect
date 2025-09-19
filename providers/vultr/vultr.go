// Package vultr implements the Vultr cloud provider detection.
package vultr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/nikhil-prabhu/clouddetect/v2/types"
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

func (v *Vultr) Identify(ctx context.Context, ch chan<- types.ProviderId, logger *zap.Logger) {
	if v.checkMetadataServer(ctx, logger) {
		ch <- v.Identifier()
		return
	}

	if v.checkVendorFile(vendorFile, logger) {
		ch <- v.Identifier()
		return
	}
}

func (v *Vultr) checkMetadataServer(ctx context.Context, logger *zap.Logger) bool {
	logger.Debug(fmt.Sprintf("Checking %s metadata using url %s", identifier, metadataURL))

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", metadataURL, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("Error creating request: %s", err))
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading response: %s", err))
		return false
	}
	defer func(Body io.ReadCloser) {
		closeErr := Body.Close()
		if closeErr != nil {
			logger.Error(fmt.Sprintf("Error closing response body: %s", closeErr))
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		logger.Error(fmt.Sprintf("Error response status code: %d", resp.StatusCode))
		return false
	}

	metadata := new(metadataResponse)
	if decodeErr := json.NewDecoder(resp.Body).Decode(metadata); decodeErr != nil {
		logger.Error(fmt.Sprintf("Error decoding response: %s", decodeErr))
		return false
	}

	return len(metadata.InstanceID) > 0
}

func (v *Vultr) checkVendorFile(file string, logger *zap.Logger) bool {
	logger.Debug(fmt.Sprintf("Checking %s vendor file %s", identifier, file))

	content, err := os.ReadFile(file)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading file: %s", err))
		return false
	}

	return strings.Contains(string(content), "Vultr")
}
