package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// PostForm 表单POST请求
func PostForm(ctx context.Context, client *http.Client, endpoint string, form url.Values, result interface{}) error {
	req, err := newRequest(ctx, http.MethodPost, endpoint, nil, form, "")
	if err != nil {
		return err
	}

	return doRequest(client, req, result)
}

func newRequest(ctx context.Context, method, endpoint string, query url.Values, form url.Values, token string) (*http.Request, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}

	if query != nil {
		u.RawQuery = query.Encode()
	}

	var body io.Reader
	contentType := ""
	if form != nil {
		body = bytes.NewBufferString(form.Encode())
		contentType = "application/x-www-form-urlencoded"
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, nil
}

func doRequest(client *http.Client, req *http.Request, result interface{}) error {
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode response failed: %w", err)
	}

	return nil
}

// GetWithCustomHeaders 带自定义Header的GET请求
func GetWithCustomHeaders(
	ctx context.Context,
	client *http.Client,
	url string,
	headers map[string]string,
	result interface{},
) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	// 添加自定义Header
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode response failed: %w", err)
	}

	return nil
}

// PostFormWithHeaders 发送带自定义Header的表单POST请求
func PostFormWithHeaders(
	ctx context.Context,
	client *http.Client,
	url string,
	form url.Values,
	headers map[string]string,
	result interface{},
) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode response failed: %w", err)
	}

	return nil
}

// PostJSONWithAuth 带认证的JSON POST请求
func PostJSONWithAuth(
	ctx context.Context,
	client *http.Client,
	endpoint string,
	body interface{},
	token string,
	result interface{},
) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal json failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode response failed: %w", err)
	}

	return nil
}

// HTTPGet 发送HTTP GET请求并将响应解析为JSON。
//
// ctx 用于控制请求的上下文，可用于超时控制和取消。
//
// url 是请求的完整URL地址。
//
// result 是用于存储JSON响应的结构体指针。
//
// 返回错误信息，可能的错误包括：
//   - 创建请求失败
//   - 发送请求失败
//   - 读取响应失败
//   - JSON解析失败
func HTTPGet(ctx context.Context, url string, result interface{}) error {
	// 创建带有上下文的请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	// 解析JSON响应
	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return nil
}
