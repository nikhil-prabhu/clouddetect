package openstack

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/nikhil-prabhu/clouddetect/logging"
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

func (o *OpenStack) Identify(ctx context.Context, ch chan<- types.ProviderId) {
	if o.checkMetadataServer(ctx) {
		ch <- o.Identifier()
		return
	}

	if o.checkProductNameFile(productNameFile) {
		ch <- o.Identifier()
		return
	}

	if o.checkChassisAssetTagFile(chassisAssetTagFile) {
		ch <- o.Identifier()
		return
	}
}

func (o *OpenStack) checkMetadataServer(ctx context.Context) bool {
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

	return resp.StatusCode == http.StatusOK
}

func (o *OpenStack) checkProductNameFile(file string) bool {
	logging.Logger.Debug(fmt.Sprintf("Checking %s product name using file %s", identifier, file))

	content, err := os.ReadFile(file)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("Error reading file: %s", err))
		return false
	}

	return slices.Contains(productNames, strings.TrimSpace(string(content)))
}

func (o *OpenStack) checkChassisAssetTagFile(file string) bool {
	logging.Logger.Debug(fmt.Sprintf("Checking %s chassis asset tag using file %s", identifier, file))

	content, err := os.ReadFile(file)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("Error reading file: %s", err))
		return false
	}

	return slices.Contains(chassisAssetTags, strings.TrimSpace(string(content)))
}
