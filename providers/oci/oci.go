// Package oci implements the Oracle Cloud Infrastructure (OCI) provider detection.
package oci

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/nikhil-prabhu/clouddetect/types"
)

const (
	metadataURL string = "http://169.254.169.254/opc/v1/instance/metadata"
	vendorFile         = "/sys/class/dmi/id/chassis_asset_tag"
	identifier         = types.Oci
)

type metadataResponse struct {
	OkeTm string `json:"oke_tm"`
}

type Oci struct{}

func (o *Oci) Identifier() types.ProviderId {
	return identifier
}

func (o *Oci) Identify(ctx context.Context, ch chan<- types.ProviderId, logger *zap.Logger) {
	if o.checkMetadataServer(ctx, logger) {
		ch <- o.Identifier()
		return
	}

	if o.checkVendorFile(vendorFile, logger) {
		ch <- o.Identifier()
		return
	}
}

func (o *Oci) checkMetadataServer(ctx context.Context, logger *zap.Logger) bool {
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

	return strings.Contains(metadata.OkeTm, "oke")
}

func (o *Oci) checkVendorFile(file string, logger *zap.Logger) bool {
	logger.Debug(fmt.Sprintf("Checking %s vendor file %s", identifier, file))

	content, err := os.ReadFile(file)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading file: %s", err))
		return false
	}

	return strings.Contains(string(content), "OracleCloud")
}
