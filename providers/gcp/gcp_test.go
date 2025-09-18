package gcp

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"go.uber.org/zap"

	"github.com/nikhil-prabhu/clouddetect/types"
)

func createTempFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "vendorfile-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, writeErr := tmpFile.WriteString(content); writeErr != nil {
		t.Fatalf("Failed to write to temp file: %v", writeErr)
	}

	if closeErr := tmpFile.Close(); closeErr != nil {
		t.Fatalf("Failed to close temp file: %v", closeErr)
	}

	return tmpFile.Name()
}

func TestIdentify(t *testing.T) {
	tests := []struct {
		name             string
		setupMocks       func()
		expectedProvider types.ProviderId
	}{
		{
			name: "Identify GCP via metadata server",
			setupMocks: func() {
				httpmock.Activate()
				httpmock.RegisterResponder("GET", metadataURL, httpmock.NewStringResponder(200, ""))
			},
			expectedProvider: identifier,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			tt.setupMocks()

			g := &Gcp{}
			ch := make(chan types.ProviderId, 1)
			logger := zap.NewNop()

			go g.Identify(context.Background(), ch, logger)

			select {
			case result := <-ch:
				if result != tt.expectedProvider {
					t.Errorf("Identify() = %v; want %v", result, tt.expectedProvider)
				}
			case <-time.After(time.Second):
				t.Error("Identify() timed out")
			}
		})
	}
}

func TestCheckMetadataServer(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	tests := []struct {
		name           string
		responseStatus int
		expectedResult bool
	}{
		{
			name:           "Successful metadata response",
			responseStatus: http.StatusOK,
			expectedResult: true,
		},
		{
			name:           "Non-OK status code",
			responseStatus: http.StatusInternalServerError,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.RegisterResponder("GET", metadataURL, httpmock.NewStringResponder(tt.responseStatus, ""))

			g := &Gcp{}
			logger := zap.NewNop()
			result := g.checkMetadataServer(context.Background(), logger)

			if result != tt.expectedResult {
				t.Errorf("checkMetadataServer() = %v; want %v", result, tt.expectedResult)
			}
		})
	}
}

func TestCheckVendorFile(t *testing.T) {
	tests := []struct {
		name           string
		fileContent    string
		expectedResult bool
	}{
		{
			name:           "Vendor file contains Google",
			fileContent:    "This machine is by Google",
			expectedResult: true,
		},
		{
			name:           "Vendor file does not contain Google",
			fileContent:    "This machine is by Another Vendor",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempFile(t, tt.fileContent)
			defer func(name string) {
				err := os.Remove(name)
				if err != nil {
					t.Fatalf("Failed to remove temp file: %v", err)
				}
			}(tmpFile)

			g := &Gcp{}
			logger := zap.NewNop()
			result := g.checkVendorFile(tmpFile, logger)

			if result != tt.expectedResult {
				t.Errorf("checkVendorFile() = %v; want %v", result, tt.expectedResult)
			}
		})
	}
}
