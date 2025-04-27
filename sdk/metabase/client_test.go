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
