package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var httpClient = http.Client{
	Timeout: time.Second * 30,
}

// SendHttpRequest 发送请求
func SendHttpRequest[T any](ctx context.Context, request *http.Request) (*T, error) {
	if request == nil {
		return nil, fmt.Errorf("request is nil")
	}
	resp, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("send request failed, err: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status code is not 200, status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed, err: %w", err)
	}
	var t T
	err = json.Unmarshal(body, &t)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response body failed, err: %w", err)
	}
	return &t, nil
}
