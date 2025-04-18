package akamai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/nikhil-prabhu/clouddetect/logging"
	"github.com/nikhil-prabhu/clouddetect/types"
)

const (
	metadataURL string = "http://169.254.169.254/v1/instance"
	tokenURL    string = "http://169.254.169.254/v1/token"
	identifier         = types.Akamai
)

type metadataResponse struct {
	Id       int    `json:"id"`
	HostUUID string `json:"host_uuid"`
}

type Akamai struct{}

func (a *Akamai) Identifier() types.ProviderId {
	return identifier
}

func (a *Akamai) getMetadata(ctx context.Context) (*metadataResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", tokenURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Metadata-Token-Expiry-Seconds", "60")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		closeErr := Body.Close()
		if closeErr != nil {
			logging.Logger.Error(fmt.Sprintf("Error closing response body: %s", closeErr))
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response status code: %d", resp.StatusCode)
	}

	token, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	req, err = http.NewRequestWithContext(ctx, "GET", metadataURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Metadata-Token", string(token))

	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		closeErr := Body.Close()
		if closeErr != nil {
			logging.Logger.Error(fmt.Sprintf("Error closing response body: %s", closeErr))
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response status code: %d", resp.StatusCode)
	}

	metadata := new(metadataResponse)
	if decodeErr := json.NewDecoder(resp.Body).Decode(metadata); decodeErr != nil {
		return nil, decodeErr
	}

	return metadata, nil
}

func (a *Akamai) Identify(ctx context.Context, ch chan<- types.ProviderId) {
	if a.checkMetadataServer(ctx) {
		ch <- a.Identifier()
		return
	}
}

func (a *Akamai) checkMetadataServer(ctx context.Context) bool {
	logging.Logger.Debug(fmt.Sprintf("Checking %s metadata using url %s", identifier, metadataURL))

	metadata, err := a.getMetadata(ctx)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("Error reading response: %s", err))
		return false
	}

	return metadata.Id > 0 && strings.TrimSpace(metadata.HostUUID) != ""
}
