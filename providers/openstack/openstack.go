// Package openstack implements the OpenStack cloud provider detection.
package openstack

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"

	"go.uber.org/zap"

	"github.com/nikhil-prabhu/clouddetect/types"
)

const (
	metadataURL         string = "http://169.254.169.254/openstack/"
	productNameFile            = "/sys/class/dmi/id/product_name"
	chassisAssetTagFile        = "/sys/class/dmi/id/chassis_asset_tag"
	identifier                 = types.OpenStack
)

var (
	productNames     = []string{"OpenStack Nova", "OpenStack Compute"}
	chassisAssetTags = []string{"HUAWEICLOUD", "OpenTelekomCloud", "SAP CCloud VM", "OpenStack Nova", "OpenStack Compute"}
)

type OpenStack struct{}

func (o *OpenStack) Identifier() types.ProviderId {
	return identifier
}

func (o *OpenStack) Identify(ctx context.Context, ch chan<- types.ProviderId, logger *zap.Logger) {
	if o.checkMetadataServer(ctx, logger) {
		ch <- o.Identifier()
		return
	}

	if o.checkProductNameFile(productNameFile, logger) {
		ch <- o.Identifier()
		return
	}

	if o.checkChassisAssetTagFile(chassisAssetTagFile, logger) {
		ch <- o.Identifier()
		return
	}
}

func (o *OpenStack) checkMetadataServer(ctx context.Context, logger *zap.Logger) bool {
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

	return resp.StatusCode == http.StatusOK
}

func (o *OpenStack) checkProductNameFile(file string, logger *zap.Logger) bool {
	logger.Debug(fmt.Sprintf("Checking %s product name using file %s", identifier, file))

	content, err := os.ReadFile(file)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading file: %s", err))
		return false
	}

	return slices.Contains(productNames, strings.TrimSpace(string(content)))
}

func (o *OpenStack) checkChassisAssetTagFile(file string, logger *zap.Logger) bool {
	logger.Debug(fmt.Sprintf("Checking %s chassis asset tag using file %s", identifier, file))

	content, err := os.ReadFile(file)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading file: %s", err))
		return false
	}

	return slices.Contains(chassisAssetTags, strings.TrimSpace(string(content)))
}
