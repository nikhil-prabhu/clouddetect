package oci

import (
	"context"
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
			name: "Identify OCI via metadata server",
			setupMocks: func() {
				httpmock.Activate()
				httpmock.RegisterResponder("GET", metadataURL, httpmock.NewJsonResponderOrPanic(200, metadataResponse{
					OkeTm: "oke-instance",
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

			o := &Oci{}
			ch := make(chan types.ProviderId, 1)

			go o.Identify(context.Background(), ch)

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
		responseBody   *metadataResponse
		expectedResult bool
	}{
		{
			name:           "Successful metadata response with OKE_TM",
			responseStatus: http.StatusOK,
			responseBody: &metadataResponse{
				OkeTm: "oke-instance",
			},
			expectedResult: true,
		},
		{
			name:           "Metadata response without OKE_TM",
			responseStatus: http.StatusOK,
			responseBody:   &metadataResponse{},
			expectedResult: false,
		},
		{
			name:           "Non-OK status code",
			responseStatus: http.StatusInternalServerError,
			responseBody:   &metadataResponse{},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.RegisterResponder("GET", metadataURL, httpmock.NewJsonResponderOrPanic(tt.responseStatus, tt.responseBody))

			o := &Oci{}
			result := o.checkMetadataServer(context.Background())

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
			name:           "Vendor file contains OracleCloud",
			fileContent:    "This machine is an OracleCloud instance",
			expectedResult: true,
		},
		{
			name:           "Vendor file does not contain OracleCloud",
			fileContent:    "This machine is by another vendor",
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

			o := &Oci{}
			result := o.checkVendorFile(tmpFile)

			if result != tt.expectedResult {
				t.Errorf("checkVendorFile() = %v; want %v", result, tt.expectedResult)
			}
		})
	}
}
