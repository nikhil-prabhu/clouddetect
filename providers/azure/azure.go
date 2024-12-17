package azure

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/nikhil-prabhu/clouddetect"
)

const (
	metadataURL string = "http://169.254.169.254/metadata/instance?api-version=2017-12-01"
	vendorFile         = "/sys/class/dmi/id/sys_vendor"
	identifier         = clouddetect.Azure
)

type compute struct {
	VmID string `json:"vmId"`
}

type metadataResponse struct {
	Compute compute `json:"compute"`
}

type Azure struct{}

func (a *Azure) Identifier() clouddetect.ProviderId {
	return identifier
}

func (a *Azure) Identify(ch chan clouddetect.ProviderId) {
	if a.checkMetadataServer() {
		ch <- identifier
		return
	}

	if a.checkVendorFile(vendorFile) {
		ch <- identifier
		return
	}
}

func (a *Azure) checkMetadataServer() bool {
	clouddetect.Logger.Debug(fmt.Sprintf("Checking %s metadata using url %s", identifier, metadataURL))

	client := &http.Client{}
	req, err := http.NewRequest("GET", metadataURL, nil)
	if err != nil {
		clouddetect.Logger.Error(fmt.Sprintf("Error creating request: %s", err))
		return false
	}
	req.Header.Add("Metadata", "true")

	resp, err := client.Do(req)
	if err != nil {
		clouddetect.Logger.Error(fmt.Sprintf("Error sending request: %s", err))
		return false
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			clouddetect.Logger.Error(fmt.Sprintf("Error closing response body: %s", err))
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		clouddetect.Logger.Error(fmt.Sprintf("Error response status code: %d", resp.StatusCode))
		return false
	}

	metadata := new(metadataResponse)
	if err = json.NewDecoder(resp.Body).Decode(metadata); err != nil {
		clouddetect.Logger.Error(fmt.Sprintf("Error decoding response: %s", err))
		return false
	}

	return len(metadata.Compute.VmID) > 0
}

func (a *Azure) checkVendorFile(file string) bool {
	clouddetect.Logger.Debug(fmt.Sprintf("Checking %s vendor file %s", identifier, file))

	content, err := os.ReadFile(file)
	if err != nil {
		clouddetect.Logger.Error(fmt.Sprintf("Error reading file: %s", err))
		return false
	}

	return strings.Contains(string(content), "Microsoft Corporation")
}
