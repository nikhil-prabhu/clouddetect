package openstack

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
	tmpFile, err := os.CreateTemp("", "testfile-*.txt")
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
			name: "Identify OpenStack via metadata server",
			setupMocks: func() {
				httpmock.Activate()
				httpmock.RegisterResponder("GET", metadataURL, httpmock.NewStringResponder(200, ""))
			},
			expectedProvider: identifier,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			defer httpmock.DeactivateAndReset()

			o := &OpenStack{}
			ch := make(chan types.ProviderId, 1)
			logger := zap.NewNop()

			go o.Identify(context.Background(), ch, logger)

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
			responseStatus: http.StatusNotFound,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.RegisterResponder("GET", metadataURL, httpmock.NewStringResponder(tt.responseStatus, ""))

			o := &OpenStack{}
			logger := zap.NewNop()
			result := o.checkMetadataServer(context.Background(), logger)

			if result != tt.expectedResult {
				t.Errorf("checkMetadataServer() = %v; want %v", result, tt.expectedResult)
			}
		})
	}
}

func TestCheckProductNameFile(t *testing.T) {
	tests := []struct {
		name           string
		fileContent    string
		expectedResult bool
	}{
		{
			name:           "Valid product name",
			fileContent:    "OpenStack Nova",
			expectedResult: true,
		},
		{
			name:           "Invalid product name",
			fileContent:    "Unknown Product",
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

			o := &OpenStack{}
			logger := zap.NewNop()
			result := o.checkProductNameFile(tmpFile, logger)

			if result != tt.expectedResult {
				t.Errorf("checkProductNameFile() = %v; want %v", result, tt.expectedResult)
			}
		})
	}
}

func TestCheckChassisAssetTagFile(t *testing.T) {
	tests := []struct {
		name           string
		fileContent    string
		expectedResult bool
	}{
		{
			name:           "Valid chassis asset tag",
			fileContent:    "HUAWEICLOUD",
			expectedResult: true,
		},
		{
			name:           "Invalid chassis asset tag",
			fileContent:    "Unknown Tag",
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

			o := &OpenStack{}
			logger := zap.NewNop()
			result := o.checkChassisAssetTagFile(tmpFile, logger)

			if result != tt.expectedResult {
				t.Errorf("checkChassisAssetTagFile() = %v; want %v", result, tt.expectedResult)
			}
		})
	}
}
