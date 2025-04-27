package metabase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type MetabaseAPIClient struct {
	Host   string
	APIKey string
	Client *http.Client
}

func (m *MetabaseAPIClient) request(ctx context.Context, path string, body any, method string) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = strings.NewReader(string(jsonBody))
	}

	url := fmt.Sprintf("%s%s", m.Host, path)
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", m.APIKey)

	resp, err := m.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

func (m *MetabaseAPIClient) Post(ctx context.Context, path string, body any) (*http.Response, error) {
	return m.request(ctx, path, body, "POST")
}

func (m *MetabaseAPIClient) Get(ctx context.Context, path string) (*http.Response, error) {
	return m.request(ctx, path, nil, "GET")
}

func (m *MetabaseAPIClient) Put(ctx context.Context, path string, body any) (*http.Response, error) {
	return m.request(ctx, path, body, "PUT")
}

func (m *MetabaseAPIClient) Delete(ctx context.Context, path string) (*http.Response, error) {
	return m.request(ctx, path, nil, "DELETE")
}
