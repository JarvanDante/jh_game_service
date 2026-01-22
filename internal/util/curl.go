package util

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//CurlPostHeader
/**
 * @desc：ONE请求方法
 * @param url 请求地址
 * @param headers 请求头 (key -> value)
 * @param data 请求body（可以是 []byte / string / map / struct）
 * @return map[string]any JSON解析后的 map
 * @return error
 * @author : Carson
 */
func CurlPostHeader(url string, headers map[string]string, data any) (map[string]any, error) {
	// 总超时 13s
	ctx, cancel := context.WithTimeout(context.Background(), 13*time.Second)
	defer cancel()

	// 处理 data -> []byte
	bodyBytes, err := encodeBody(data)
	if err != nil {
		return nil, fmt.Errorf("encode body failed: %w", err)
	}

	// 构造 HTTP Client（连接超时 5s + 忽略证书校验）
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 等价 CURLOPT_SSL_VERIFYPEER=false
			DialContext: (&net.Dialer{
				Timeout: 5 * time.Second, // 连接超时 5s
			}).DialContext,
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("new request failed: %w", err)
	}

	// 设置 headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 如果没传 Content-Type，默认给 JSON
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// 发起请求
	resp, err := client.Do(req)
	if err != nil {
		// 网络错误 / 超时
		return nil, fmt.Errorf("network request failed: %s, err=%w", url, err)
	}
	defer resp.Body.Close()

	// 读取 body
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %s, err=%w", url, err)
	}

	// 状态码非 2xx 也算异常
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http status error: %s, status=%d, body=%s", url, resp.StatusCode, truncate(string(rawBody), 500))
	}

	// JSON 解析
	var result map[string]any
	if err := json.Unmarshal(rawBody, &result); err != nil {
		return nil, fmt.Errorf("json parse failed: %s, body=%s, err=%w", url, truncate(string(rawBody), 500), err)
	}

	// 空对象判定
	if len(result) == 0 {
		return nil, fmt.Errorf("empty json response: %s, body=%s", url, truncate(string(rawBody), 500))
	}

	return result, nil
}

// encodeBody 是我们自己写的辅助函数（Go 没有内置）
func encodeBody(data any) ([]byte, error) {
	switch v := data.(type) {
	case nil:
		return []byte{}, nil
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return json.Marshal(v) // struct/map 自动转 JSON
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

//type SendResp struct {
//	HTTPCode int    `json:"http_code"`
//	Body     string `json:"body"`
//}

//SendRequest
/**
 * @desc：IG_JILI请求方法
 * @param urlStr
 * @param params
 * @return *SendResp
 * @return error
 * @author : Carson
 */
func SendRequest(urlStr string, params map[string]string) (map[string]interface{}, error) {
	// 连接超时10秒，总超时30秒
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// http_build_query 等价：url.Values{}.Encode()
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	postData := values.Encode() // a=1&b=2

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, strings.NewReader(postData))
	if err != nil {
		return nil, fmt.Errorf("new request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Transport: &http.Transport{
			// CURLOPT_SSL_VERIFYPEER=false
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},

			// CURLOPT_CONNECTTIMEOUT=10
			DialContext: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).DialContext,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("curl request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	return map[string]interface{}{
		"http_code": resp.StatusCode,
		"body":      bodyBytes,
	}, nil
	//return &SendResp{
	//	HTTPCode: resp.StatusCode,
	//	Body:     string(bodyBytes),
	//}, nil
}
