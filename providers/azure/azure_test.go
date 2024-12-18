package azure

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"

	"github.com/nikhil-prabhu/clouddetect/types"
)

func TestIdentifier(t *testing.T) {
	a := &Azure{}
	if a.Identifier() != identifier {
		t.Errorf("identifier() = %v; want %v", a.Identifier(), identifier)
	}
}

func TestIdentify(t *testing.T) {
	tests := []struct {
		name             string
		setupMocks       func()
		expectedProvider types.ProviderId
	}{
		{
			name: "Identify Azure via metadata server",
			setupMocks: func() {
				httpmock.RegisterResponder("GET", metadataURL, httpmock.NewJsonResponderOrPanic(200, metadataResponse{
					Compute: compute{VmID: "vm-12345"},
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

			a := &Azure{}
			ch := make(chan types.ProviderId, 1)

			go a.Identify(ch)

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
		responseBody   string
		responseStatus int
		expectedResult bool
	}{
		{
			name:           "Valid metadata response",
			responseBody:   `{"compute": {"vmId": "vm-12345"}}`,
			responseStatus: http.StatusOK,
			expectedResult: true,
		},
		{
			name:           "Invalid JSON response",
			responseBody:   `{"compute": {}}`,
			responseStatus: http.StatusOK,
			expectedResult: false,
		},
		{
			name:           "Non-OK status code",
			responseBody:   "",
			responseStatus: http.StatusInternalServerError,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.RegisterResponder("GET", metadataURL, httpmock.NewStringResponder(tt.responseStatus, tt.responseBody))

			a := &Azure{}
			result := a.checkMetadataServer()

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
			name:           "Vendor file contains Microsoft Corporation",
			fileContent:    "This machine is by Microsoft Corporation",
			expectedResult: true,
		},
		{
			name:           "Vendor file does not contain Microsoft Corporation",
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

			a := &Azure{}
			result := a.checkVendorFile(tmpFile)

			if result != tt.expectedResult {
				t.Errorf("checkVendorFile() = %v; want %v", result, tt.expectedResult)
			}
		})
	}
}

func TestCheckVendorFile_FileNotFound(t *testing.T) {
	a := &Azure{}
	result := a.checkVendorFile("/path/to/nonexistent/file")

	if result {
		t.Errorf("Expected checkVendorFile() to return false for nonexistent file")
	}
}

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
