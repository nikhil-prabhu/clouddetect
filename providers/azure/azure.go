// Package azure implements detection for Microsoft Azure cloud service provider.
package azure

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
	metadataURL string = "http://169.254.169.254/metadata/instance?api-version=2017-12-01"
	vendorFile         = "/sys/class/dmi/id/sys_vendor"
	identifier         = types.Azure
)

type compute struct {
	VMID string `json:"vmId"`
}

type metadataResponse struct {
	Compute compute `json:"compute"`
}

type Azure struct{}

func (a *Azure) Identifier() types.ProviderId {
	return identifier
}

func (a *Azure) Identify(ctx context.Context, ch chan<- types.ProviderId, logger *zap.Logger) {
	if a.checkMetadataServer(ctx, logger) {
		ch <- a.Identifier()
		return
	}

	if a.checkVendorFile(vendorFile, logger) {
		ch <- a.Identifier()
		return
	}
}

func (a *Azure) checkMetadataServer(ctx context.Context, logger *zap.Logger) bool {
	logger.Debug(fmt.Sprintf("Checking %s metadata using url %s", identifier, metadataURL))

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", metadataURL, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("Error creating request: %s", err))
		return false
	}
	req.Header.Add("Metadata", "true")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("Error sending request: %s", err))
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

	return len(metadata.Compute.VMID) > 0
}

func (a *Azure) checkVendorFile(file string, logger *zap.Logger) bool {
	logger.Debug(fmt.Sprintf("Checking %s vendor file %s", identifier, file))

	content, err := os.ReadFile(file)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading file: %s", err))
		return false
	}

	return strings.Contains(string(content), "Microsoft Corporation")
}
