// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package metabase

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

type mockHTTPClient struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
	LastRequest   *http.Request
}

func (m *mockHTTPClient) RoundTrip(req *http.Request) (*http.Response, error) {
	m.LastRequest = req
	return m.RoundTripFunc(req)
}

func newTestClient(mock *mockHTTPClient) *MetabaseAPIClient {
	return &MetabaseAPIClient{
		Host:   "http://localhost:3000",
		APIKey: "test-key",
		Client: &http.Client{Transport: mock},
	}
}

func assertRequest(t *testing.T, req *http.Request, expectedMethod, expectedURL string, expectedHeaders map[string]string, expectedBody interface{}) {
	t.Helper()

	if req == nil {
		t.Fatalf("expected a request to be made")
	}

	if req.Method != expectedMethod {
		t.Errorf("expected method %s, got %s", expectedMethod, req.Method)
	}

	if req.URL.String() != expectedURL {
		t.Errorf("expected URL %q, got %q", expectedURL, req.URL.String())
	}

	for key, expectedValue := range expectedHeaders {
		if actualValue := req.Header.Get(key); actualValue != expectedValue {
			t.Errorf("expected header %q: %q, got %q", key, expectedValue, actualValue)
		}
	}

	if expectedBody != nil {
		requestBodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		if len(requestBodyBytes) == 0 && expectedBody != nil {
			t.Errorf("expected non-empty body, got empty")
			return
		}

		expectedBodyBytes, err := json.Marshal(expectedBody)
		if err != nil {
			t.Fatalf("failed to marshal expected body: %v", err)
		}

		if !bytes.Equal(requestBodyBytes, expectedBodyBytes) {
			t.Errorf("expected body %q, got %q", string(expectedBodyBytes), string(requestBodyBytes))
		}
	} else if req.Body != nil {
		requestBodyBytes, _ := io.ReadAll(req.Body)
		if len(requestBodyBytes) > 0 {
			t.Errorf("expected empty body, got %q", string(requestBodyBytes))
		}
	}
}

func TestMetabaseAPIClient_Post(t *testing.T) {
	mockClient := &mockHTTPClient{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(bytes.NewBufferString(`{"id": 123}`)),
			}, nil
		},
	}
	client := newTestClient(mockClient)
	expectedBody := map[string]string{"name": "test"}
	expectedHeaders := map[string]string{"Content-Type": "application/json", "x-api-key": "test-key"}
	expectedURL := "http://localhost:3000/api/items"

	_, err := client.Post(context.Background(), "/api/items", expectedBody)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}

	assertRequest(t, mockClient.LastRequest, http.MethodPost, expectedURL, expectedHeaders, expectedBody)
}

func TestMetabaseAPIClient_Get(t *testing.T) {
	mockClient := &mockHTTPClient{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"name": "test item"}`)),
			}, nil
		},
	}
	client := newTestClient(mockClient)
	expectedHeaders := map[string]string{"Content-Type": "application/json", "x-api-key": "test-key"}
	expectedURL := "http://localhost:3000/api/items/1"

	_, err := client.Get(context.Background(), "/api/items/1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	assertRequest(t, mockClient.LastRequest, http.MethodGet, expectedURL, expectedHeaders, nil)
}

func TestMetabaseAPIClient_Put(t *testing.T) {
	mockClient := &mockHTTPClient{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"id": 1, "name": "updated item"}`)),
			}, nil
		},
	}
	client := newTestClient(mockClient)
	expectedBody := map[string]string{"name": "updated"}
	expectedHeaders := map[string]string{"Content-Type": "application/json", "x-api-key": "test-key"}
	expectedURL := "http://localhost:3000/api/items/1"

	_, err := client.Put(context.Background(), "/api/items/1", expectedBody)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	assertRequest(t, mockClient.LastRequest, http.MethodPut, expectedURL, expectedHeaders, expectedBody)
}

func TestMetabaseAPIClient_Delete(t *testing.T) {
	mockClient := &mockHTTPClient{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Body:       io.NopCloser(bytes.NewBufferString("")),
			}, nil
		},
	}
	client := newTestClient(mockClient)
	expectedHeaders := map[string]string{"Content-Type": "application/json", "x-api-key": "test-key"}
	expectedURL := "http://localhost:3000/api/items/1"

	_, err := client.Delete(context.Background(), "/api/items/1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	assertRequest(t, mockClient.LastRequest, http.MethodDelete, expectedURL, expectedHeaders, nil)
}

func TestMetabaseAPIClient_ErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		path         string
		expectedType interface{}
	}{
		{
			name:         "NotFoundError",
			statusCode:   http.StatusNotFound,
			path:         "/api/user/123",
			expectedType: &NotFoundError{},
		},
		{
			name:         "ConflictError",
			statusCode:   http.StatusConflict,
			path:         "/api/user",
			expectedType: &ConflictError{},
		},
		{
			name:         "BadRequestError",
			statusCode:   http.StatusBadRequest,
			path:         "/api/user",
			expectedType: &BadRequestError{},
		},
		{
			name:         "UnauthorizedError",
			statusCode:   http.StatusUnauthorized,
			path:         "/api/user",
			expectedType: &UnauthorizedError{},
		},
		{
			name:         "ForbiddenError",
			statusCode:   http.StatusForbidden,
			path:         "/api/user/123",
			expectedType: &ForbiddenError{},
		},
		{
			name:         "GenericError",
			statusCode:   http.StatusInternalServerError,
			path:         "/api/user",
			expectedType: &BaseError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockHTTPClient{
				RoundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: tt.statusCode,
						Body:       io.NopCloser(bytes.NewBufferString("Error message")),
					}, nil
				},
			}
			client := newTestClient(mockClient)

			_, err := client.Get(context.Background(), tt.path)

			if err == nil {
				t.Fatalf("Expected error but got nil")
			}

			// Check if the error is of the expected type
			switch tt.expectedType.(type) {
			case *NotFoundError:
				if _, ok := err.(*NotFoundError); !ok {
					t.Errorf("Expected NotFoundError but got %T", err)
				}
			case *ConflictError:
				if _, ok := err.(*ConflictError); !ok {
					t.Errorf("Expected ConflictError but got %T", err)
				}
			case *BadRequestError:
				if _, ok := err.(*BadRequestError); !ok {
					t.Errorf("Expected BadRequestError but got %T", err)
				}
			case *UnauthorizedError:
				if _, ok := err.(*UnauthorizedError); !ok {
					t.Errorf("Expected UnauthorizedError but got %T", err)
				}
			case *ForbiddenError:
				if _, ok := err.(*ForbiddenError); !ok {
					t.Errorf("Expected ForbiddenError but got %T", err)
				}
			case *BaseError:
				if _, ok := err.(*BaseError); !ok {
					t.Errorf("Expected BaseError but got %T", err)
				}
			}
		})
	}
}
