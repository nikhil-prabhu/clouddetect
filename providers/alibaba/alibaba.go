// Package alibaba implements the Alibaba Cloud provider detection.
package alibaba

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/nikhil-prabhu/clouddetect/v2/types"
)

const (
	metadataURL string = "http://100.100.100.200/latest/meta-data/latest/meta-data/instance/virtualization-solution"
	vendorFile         = "/sys/class/dmi/id/product_name"
	identifier         = types.Alibaba
)

type Alibaba struct{}

func (a *Alibaba) Identifier() types.ProviderId {
	return identifier
}

func (a *Alibaba) Identify(ctx context.Context, ch chan<- types.ProviderId, logger *zap.Logger) {
	if a.checkMetadataServer(ctx, logger) {
		ch <- a.Identifier()
		return
	}

	if a.checkVendorFile(vendorFile, logger) {
		ch <- a.Identifier()
		return
	}
}

func (a *Alibaba) checkMetadataServer(ctx context.Context, logger *zap.Logger) bool {
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

	text, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading response body: %s", err))
		return false
	}

	return strings.Contains(string(text), "ECS Virt")
}

func (a *Alibaba) checkVendorFile(vendorFile string, logger *zap.Logger) bool {
	logger.Debug(fmt.Sprintf("Checking %s vendor file %s", identifier, vendorFile))

	content, err := os.ReadFile(vendorFile)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading file: %s", err))
		return false
	}

	return strings.Contains(string(content), "Alibaba Cloud ECS")
}
