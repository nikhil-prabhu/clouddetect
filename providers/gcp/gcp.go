package gcp

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/nikhil-prabhu/clouddetect/logging"
	"github.com/nikhil-prabhu/clouddetect/types"
)

const (
	metadataURL string = "http://metadata.google.internal/computeMetadata/v1/instance/tags"
	vendorFile         = "/sys/class/dmi/id/product_name"
	identifier         = types.Gcp
)

type Gcp struct{}

func (g *Gcp) Identifier() types.ProviderId {
	return identifier
}

func (g *Gcp) Identify(ch chan<- types.ProviderId) {
	if g.checkMetadataServer() {
		ch <- g.Identifier()
		return
	}

	if g.checkVendorFile(vendorFile) {
		ch <- g.Identifier()
		return
	}
}

func (g *Gcp) checkMetadataServer() bool {
	logging.Logger.Debug(fmt.Sprintf("Checking %s metadata using url %s", identifier, metadataURL))

	client := &http.Client{}
	req, err := http.NewRequest("GET", metadataURL, nil)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("Error creating request: %s", err))
		return false
	}
	req.Header.Add("Metadata-Flavor", "Google")

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

	return resp.StatusCode == http.StatusOK
}

func (g *Gcp) checkVendorFile(file string) bool {
	logging.Logger.Debug(fmt.Sprintf("Checking %s vendor file %s", identifier, file))

	content, err := os.ReadFile(file)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("Error reading file: %s", err))
		return false
	}

	return strings.Contains(string(content), "Google")
}