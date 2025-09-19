// Package akamai implements the Akamai provider detection.
package akamai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/nikhil-prabhu/clouddetect/v2/types"
)

const (
	metadataURL string = "http://169.254.169.254/v1/instance"
	tokenURL    string = "http://169.254.169.254/v1/token"
	identifier         = types.Akamai
)

type metadataResponse struct {
	ID       int    `json:"id"`
	HostUUID string `json:"host_uuid"`
}

type Akamai struct{}

func (a *Akamai) Identifier() types.ProviderId {
	return identifier
}

func (a *Akamai) getMetadata(ctx context.Context, logger *zap.Logger) (*metadataResponse, error) {
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
			logger.Error(fmt.Sprintf("Error closing response body: %s", closeErr))
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
			logger.Error(fmt.Sprintf("Error closing response body: %s", closeErr))
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

func (a *Akamai) Identify(ctx context.Context, ch chan<- types.ProviderId, logger *zap.Logger) {
	if a.checkMetadataServer(ctx, logger) {
		ch <- a.Identifier()
		return
	}
}

func (a *Akamai) checkMetadataServer(ctx context.Context, logger *zap.Logger) bool {
	logger.Debug(fmt.Sprintf("Checking %s metadata using url %s", identifier, metadataURL))

	metadata, err := a.getMetadata(ctx, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading response: %s", err))
		return false
	}

	return metadata.ID > 0 && strings.TrimSpace(metadata.HostUUID) != ""
}
