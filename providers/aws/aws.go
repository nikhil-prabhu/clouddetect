// Package aws implements detection for Amazon Web Services (AWS) cloud service provider.
package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/nikhil-prabhu/clouddetect/types"
	"go.uber.org/zap"
)

const (
	metadataURL        string = "http://169.254.169.254/latest/dynamic/instance-identity/document"
	tokenURL           string = "http://169.254.169.254/latest/api/token"
	productVersionFile        = "/sys/class/dmi/id/product_version"
	biosVendorFile            = "/sys/class/dmi/id/bios_vendor"
	identifier                = types.Aws
)

type metadataResponse struct {
	ImageID    string `json:"imageId"`
	InstanceID string `json:"instanceId"`
}

type Aws struct{}

func (a *Aws) Identifier() types.ProviderId {
	return identifier
}

func (a *Aws) getMetadataIMDSv1(ctx context.Context, logger *zap.Logger) (*metadataResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", metadataURL, nil)
	if err != nil {
		return nil, err
	}

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

	metadata := new(metadataResponse)
	if decodeErr := json.NewDecoder(resp.Body).Decode(metadata); decodeErr != nil {
		return nil, decodeErr
	}

	return metadata, nil
}

func (a *Aws) getMetadataIMDSv2(ctx context.Context, logger *zap.Logger) (*metadataResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", tokenURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-aws-ec2-metadata-token-ttl-seconds", "60")

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
	req.Header.Add("X-aws-ec2-metadata-token", string(token))

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

func (a *Aws) Identify(ctx context.Context, ch chan<- types.ProviderId, logger *zap.Logger) {
	if a.checkMetadataServerV2(ctx, logger) {
		ch <- a.Identifier()
		return
	}

	if a.checkMetadataServerV1(ctx, logger) {
		ch <- a.Identifier()
		return
	}

	if a.checkProductVersionFile(productVersionFile, logger) {
		ch <- a.Identifier()
		return
	}

	if a.checkBiosVendorFile(biosVendorFile, logger) {
		ch <- a.Identifier()
		return
	}
}

func (a *Aws) checkMetadataServerV2(ctx context.Context, logger *zap.Logger) bool {
	logger.Debug(fmt.Sprintf("Checking %s metadata using url %s", identifier, metadataURL))

	metadata, err := a.getMetadataIMDSv2(ctx, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading response: %s", err))
		return false
	}

	return strings.HasPrefix(metadata.ImageID, "ami-") && strings.HasPrefix(metadata.InstanceID, "i-")
}

func (a *Aws) checkMetadataServerV1(ctx context.Context, logger *zap.Logger) bool {
	logger.Debug(fmt.Sprintf("Checking %s metadata using url %s", identifier, metadataURL))

	metadata, err := a.getMetadataIMDSv1(ctx, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading response: %s", err))
		return false
	}

	return strings.HasPrefix(metadata.ImageID, "ami-") && strings.HasPrefix(metadata.InstanceID, "i-")
}

func (a *Aws) checkProductVersionFile(file string, logger *zap.Logger) bool {
	logger.Debug(fmt.Sprintf("Checking %s product version file %s", identifier, file))

	content, err := os.ReadFile(file)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading file: %s", err))
		return false
	}

	return strings.Contains(strings.ToLower(string(content)), "amazon")
}

func (a *Aws) checkBiosVendorFile(file string, logger *zap.Logger) bool {
	logger.Debug(fmt.Sprintf("Checking %s bios vendor file %s", identifier, file))

	content, err := os.ReadFile(file)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading file: %s", err))
		return false
	}

	return strings.Contains(strings.ToLower(string(content)), "amazon")
}
