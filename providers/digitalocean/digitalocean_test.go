package digitalocean

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"

	"github.com/nikhil-prabhu/clouddetect/types"
)

func createTempFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "vendorfile-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
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
			name: "Identify DigitalOcean via metadata server",
			setupMocks: func() {
				httpmock.Activate()
				httpmock.RegisterResponder("GET", metadataURL, httpmock.NewJsonResponderOrPanic(200, metadataResponse{
					DropletID: 12345678,
				}))
			},
			expectedProvider: identifier,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			tt.setupMocks()

			d := &DigitalOcean{}
			ch := make(chan types.ProviderId, 1)

			go d.Identify(ch)

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
		responseBody   *metadataResponse
		responseStatus int
		expectedResult bool
	}{
		{
			name: "Valid metadata response",
			responseBody: &metadataResponse{
				DropletID: 12345678,
			},
			responseStatus: http.StatusOK,
			expectedResult: true,
		},
		{
			name: "Invalid JSON response",
			responseBody: &metadataResponse{
				DropletID: 0,
			},
			responseStatus: http.StatusOK,
			expectedResult: false,
		},
		{
			name:           "Non-OK status code",
			responseBody:   nil,
			responseStatus: http.StatusInternalServerError,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.RegisterResponder("GET", metadataURL, httpmock.NewJsonResponderOrPanic(tt.responseStatus, tt.responseBody))

			d := &DigitalOcean{}
			result := d.checkMetadataServer()

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
			name:           "Vendor file contains DigitalOcean",
			fileContent:    "This machine is by DigitalOcean",
			expectedResult: true,
		},
		{
			name:           "Vendor file does not contain DigitalOcean",
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

			d := &DigitalOcean{}
			result := d.checkVendorFile(tmpFile)

			if result != tt.expectedResult {
				t.Errorf("checkVendorFile() = %v; want %v", result, tt.expectedResult)
			}
		})
	}
}
