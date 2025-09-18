package akamai

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"go.uber.org/zap"

	"github.com/nikhil-prabhu/clouddetect/types"
)

func TestIdentifier(t *testing.T) {
	a := &Akamai{}
	if a.Identifier() != identifier {
		t.Errorf("identifier() = %v; want %v", a.Identifier(), identifier)
	}
}

func TestIdentify(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func()
		expectedResult types.ProviderId
	}{
		{
			name: "Token retrieval succeeds",
			setupMock: func() {
				httpmock.RegisterResponder("GET", tokenURL, httpmock.NewStringResponder(200, "test-token"))
				httpmock.RegisterResponder("GET", metadataURL,
					httpmock.NewJsonResponderOrPanic(200, metadataResponse{
						ID:       123,
						HostUUID: "123abc",
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

			a := &Akamai{}
			ch := make(chan types.ProviderId)
			logger := zap.NewNop()

			go a.Identify(context.Background(), ch, logger)

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

func TestGetMetadata(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock token and metadata responses
	httpmock.RegisterResponder("GET", tokenURL, httpmock.NewStringResponder(200, "test-token"))
	httpmock.RegisterResponder("GET", metadataURL,
		func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("Metadata-Token") != "test-token" {
				return httpmock.NewStringResponse(403, ""), nil
			}
			return httpmock.NewJsonResponse(200, metadataResponse{
				ID:       123,
				HostUUID: "123abc",
			})
		},
	)

	a := &Akamai{}
	logger := zap.NewNop()
	metadata, err := a.getMetadata(context.Background(), logger)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if metadata.ID != 123 || metadata.HostUUID != "123abc" {
		t.Errorf("Incorrect metadata: %v", metadata)
	}
}

func TestCheckMetadataServer(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", tokenURL, httpmock.NewStringResponder(200, "test-token"))
	httpmock.RegisterResponder("GET", metadataURL,
		httpmock.NewJsonResponderOrPanic(200, metadataResponse{
			ID:       123,
			HostUUID: "123abc",
		}))

	a := &Akamai{}
	logger := zap.NewNop()
	if !a.checkMetadataServer(context.Background(), logger) {
		t.Error("Expected checkMetadataServer to return true")
	}
}
