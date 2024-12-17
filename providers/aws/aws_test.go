package aws

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/nikhil-prabhu/clouddetect"
)

func TestIdentifier(t *testing.T) {
	a := &Aws{}
	if a.Identifier() != identifier {
		t.Errorf("identifier() = %v; want %v", a.Identifier(), identifier)
	}
}

func TestIdentify(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func()
		expectedResult clouddetect.ProviderId
	}{
		{
			name: "IMDSv2 succeeds",
			setupMock: func() {
				httpmock.RegisterResponder("GET", tokenURL, httpmock.NewStringResponder(200, "test-token"))
				httpmock.RegisterResponder("GET", metadataURL,
					httpmock.NewJsonResponderOrPanic(200, metadataResponse{
						ImageId:    "ami-123",
						InstanceId: "i-123",
					}))
			},
			expectedResult: identifier,
		},
		{
			name: "IMDSv1 succeeds",
			setupMock: func() {
				httpmock.RegisterResponder("GET", metadataURL,
					httpmock.NewJsonResponderOrPanic(200, metadataResponse{
						ImageId:    "ami-123",
						InstanceId: "i-123",
					}))
			},
			expectedResult: identifier,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			tt.setupMock()

			a := &Aws{}
			ch := make(chan clouddetect.ProviderId)

			go a.Identify(ch)

			select {
			case result := <-ch:
				if result != tt.expectedResult {
					t.Errorf("Identify() = %v; want %v", result, tt.expectedResult)
				}
			case <-time.After(time.Second):
				t.Error("Identify() timed out")
			}
		})
	}
}

func TestGetMetadataIMDSv1(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	mockResponse := metadataResponse{
		ImageId:    "ami-12345678",
		InstanceId: "i-0123456789abcdef0",
	}
	httpmock.RegisterResponder("GET", metadataURL, httpmock.NewJsonResponderOrPanic(200, mockResponse))

	a := &Aws{}
	metadata, err := a.getMetadataIMDSv1()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if metadata.ImageId != "ami-12345678" || metadata.InstanceId != "i-0123456789abcdef0" {
		t.Errorf("Incorrect metadata: %v", metadata)
	}
}

func TestGetMetadataIMDSv2(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock IMDSv2 token and metadata responses
	httpmock.RegisterResponder("GET", tokenURL, httpmock.NewStringResponder(200, "test-token"))
	httpmock.RegisterResponder("GET", metadataURL,
		func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("X-aws-ec2-metadata-token") != "test-token" {
				return httpmock.NewStringResponse(403, ""), nil
			}
			return httpmock.NewJsonResponse(200, metadataResponse{
				ImageId:    "ami-12345678",
				InstanceId: "i-0123456789abcdef0",
			})
		},
	)

	a := &Aws{}
	metadata, err := a.getMetadataIMDSv2()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if metadata.ImageId != "ami-12345678" || metadata.InstanceId != "i-0123456789abcdef0" {
		t.Errorf("Incorrect metadata: %v", metadata)
	}
}

func TestCheckMetadataServerV1(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", metadataURL, httpmock.NewJsonResponderOrPanic(200, metadataResponse{
		ImageId:    "ami-12345678",
		InstanceId: "i-0123456789abcdef0",
	}))

	a := &Aws{}
	if !a.checkMetadataServerV1() {
		t.Error("Expected checkMetadataServerV1 to return true")
	}
}

func TestCheckMetadataServerV2(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", tokenURL, httpmock.NewStringResponder(200, "test-token"))
	httpmock.RegisterResponder("GET", metadataURL,
		httpmock.NewJsonResponderOrPanic(200, metadataResponse{
			ImageId:    "ami-12345678",
			InstanceId: "i-0123456789abcdef0",
		}))

	a := &Aws{}
	if !a.checkMetadataServerV2() {
		t.Error("Expected checkMetadataServerV2 to return true")
	}
}

func TestCheckProductVersionFile(t *testing.T) {
	content := "amazon product version"
	tmpFile := createTempFile(t, content)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatalf("Failed to remove temp file: %v", err)
		}
	}(tmpFile)

	a := &Aws{}
	if !a.checkProductVersionFile(tmpFile) {
		t.Errorf("Expected checkProductVersionFile to return true")
	}
}

func TestCheckBiosVendorFile(t *testing.T) {
	content := "amazon bios vendor"
	tmpFile := createTempFile(t, content)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatalf("Failed to remove temp file: %v", err)
		}
	}(tmpFile)

	a := &Aws{}
	if !a.checkBiosVendorFile(tmpFile) {
		t.Errorf("Expected checkBiosVendorFile to return true")
	}
}

func createTempFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "testfile-*.txt")
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
