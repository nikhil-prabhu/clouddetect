package alibaba

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"go.uber.org/zap"

	"github.com/nikhil-prabhu/clouddetect/logging"
	"github.com/nikhil-prabhu/clouddetect/types"
)

func TestIdentifier(t *testing.T) {
	a := &Alibaba{}
	if a.Identifier() != identifier {
		t.Errorf("identifier() = %v; want %v", a.Identifier(), identifier)
	}
}

func TestIdentify(t *testing.T) {
	tests := []struct {
		name           string
		responder      httpmock.Responder
		expectedResult types.ProviderId
	}{
		{
			name:           "Metadata server reachable",
			responder:      httpmock.NewStringResponder(200, "ECS Virt"),
			expectedResult: identifier,
		},
		{
			name:           "Metadata server unreachable",
			responder:      httpmock.NewErrorResponder(errors.New("error")),
			expectedResult: types.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder("GET", metadataURL, tt.responder)

			a := &Alibaba{}
			ch := make(chan types.ProviderId)

			// Start Identify in a goroutine
			go a.Identify(ch)

			// Close the channel after a timeout to simulate the failure case
			go func() {
				time.Sleep(1 * time.Millisecond)
				close(ch)
			}()

			// Wait for a result or channel close
			result, ok := <-ch
			if !ok {
				// If the channel is closed without sending a value, handle failure case
				result = types.Unknown
			}

			if result != tt.expectedResult {
				t.Errorf("Identify() = %v; want %v", result, tt.expectedResult)
			}
		})
	}
}

func TestCheckMetadataServer(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		expectPass bool
	}{
		{
			name:       "Success",
			statusCode: 200,
			response:   "ECS Virt",
			expectPass: true,
		},
		{
			name:       "Failure",
			statusCode: 404,
			response:   "",
			expectPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			// Register a mock response
			httpmock.RegisterResponder("GET", metadataURL, httpmock.NewStringResponder(tt.statusCode, tt.response))

			a := &Alibaba{}
			result := a.checkMetadataServer()

			if result != tt.expectPass {
				t.Errorf("checkMetadataServer() = %v; want %v", result, tt.expectPass)
			}
		})
	}
}

// Function to create a temporary file with given content
func createTempFile(t *testing.T, content string) string {
	t.Helper() // Marks this function as a test helper
	tmpFile, err := os.CreateTemp("", "vendorfile-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Write content to the temp file
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Close the file to flush the content
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	return tmpFile.Name() // Return the file name (path)
}

// Unit test for checkVendorFile
func TestCheckVendorFile(t *testing.T) {
	a := &Alibaba{}

	t.Run("FileContainsAlibabaCloudECS", func(t *testing.T) {
		// Arrange
		content := "This is an Alibaba Cloud ECS instance."
		tempFile := createTempFile(t, content)
		defer func(name string) {
			err := os.Remove(name)
			if err != nil {
				logging.Logger.Error("Error removing temp file", zap.Error(err))
			}
		}(tempFile) // Ensure cleanup

		// Act
		result := a.checkVendorFile(tempFile)

		// Assert
		if !result {
			t.Errorf("Expected true, got false for content containing 'Alibaba Cloud ECS'")
		}
	})

	t.Run("FileDoesNotContainAlibabaCloudECS", func(t *testing.T) {
		// Arrange
		content := "This is some other cloud provider."
		tempFile := createTempFile(t, content)
		defer func(name string) {
			err := os.Remove(name)
			if err != nil {
				logging.Logger.Error("Error removing temp file", zap.Error(err))
			}
		}(tempFile) // Ensure cleanup

		// Act
		result := a.checkVendorFile(tempFile)

		// Assert
		if result {
			t.Errorf("Expected false, got true for content not containing 'Alibaba Cloud ECS'")
		}
	})

	t.Run("FileDoesNotExist", func(t *testing.T) {
		// Act
		result := a.checkVendorFile("/path/to/nonexistent/file")

		// Assert
		if result {
			t.Errorf("Expected false, got true for nonexistent file")
		}
	})
}
