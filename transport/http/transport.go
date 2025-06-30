package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"

	"math/rand"
)

var (
	randomHosts = []string{
		// "www.google.com",
		// "api.cloudflare.com",
		// "cdn.aws.com",
		// "images.microsoft.com",
		"localhost:8080",
	}

	randomUserAgents = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		"AppleWebKit/537.36 (KHTML, like Gecko)",
		"Chrome/91.0.4472.124 Safari/537.36",
		"Edge/91.0.864.59",
	}

	randomPaths = []string{
		// "/api/v1/users",
		// "/graphql",
		// "/rest/data",
		// "/wp-json/wp/v2/posts",
		"/",
	}
)

type HTTPCamouflage struct {
}

func (h *HTTPCamouflage) getRandomElement(list []string) string {
	return list[rand.Intn(len(list))]
}

func (h *HTTPCamouflage) Wrap(buf []byte, proto string) []byte {
	buf = append([]byte(proto+":"), buf...)
	host := h.getRandomElement(randomHosts)
	ua := h.getRandomElement(randomUserAgents)
	path := h.getRandomElement(randomPaths)

	request := fmt.Sprintf(
		"POST %s HTTP/1.1\r\n"+
			"Host: %s\r\n"+
			"User-Agent: %s\r\n"+
			"Content-Type: application/octet-stream\r\n"+
			"Content-Length: %d\r\n\r\n",
		path, host, ua, len(buf))

	return append([]byte(request), buf...)
}

func (h *HTTPCamouflage) Unwrap(data []byte) (string, []byte, error) {
	// Parse the HTTP request
	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(data)))
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse HTTP request: %v", err)
	}
	defer req.Body.Close()

	// Read the complete body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read request body: %v", err)
	}

	// Split protocol and data
	parts := bytes.SplitN(body, []byte(":"), 2)
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("invalid message format, expected 'protocol:data'")
	}

	return string(parts[0]), parts[1], nil
}

func (h *HTTPCamouflage) WrapResponse(data []byte, protocol string, statusCode int) []byte {
	// اعتبارسنجی ورودی
	if len(protocol) == 0 || len(data) == 0 {
		return nil
	}

	// انتخاب تصادفی
	host := h.getRandomElement(randomHosts)
	ua := h.getRandomElement(randomUserAgents)

	// محاسبه طول محتوا
	content := fmt.Sprintf("%s:%s", protocol, data)
	contentLength := len(content)

	// ساخت پاسخ
	response := fmt.Sprintf(
		"HTTP/1.1 %d %s\r\n"+
			"Host: %s\r\n"+
			"User-Agent: %s\r\n"+
			"Content-Type: application/octet-stream\r\n"+
			"Content-Length: %d\r\n"+
			"X-Content-Type-Options: nosniff\r\n"+
			"Connection: close\r\n\r\n"+
			"%s",
		statusCode,
		http.StatusText(statusCode),
		host,
		ua,
		contentLength,
		content)

	return []byte(response)
}

func (h *HTTPCamouflage) UnwrapResponse(data []byte) (string, []byte, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	resp, err := http.ReadResponse(reader, nil)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse HTTP response: %v", err)
	}
	defer resp.Body.Close()

	// بررسی اندازه پاسخ
	if resp.ContentLength > 10*1024*1024 { // 10MB max
		return "", nil, fmt.Errorf("response too large")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// بررسی تطابق Content-Length
	if resp.ContentLength > 0 && int64(len(body)) != resp.ContentLength {
		return "", nil, fmt.Errorf("content length mismatch")
	}

	// بررسی وجود پروتکل
	if !bytes.Contains(body, []byte(":")) {
		return "", body, nil
	}

	parts := bytes.SplitN(body, []byte(":"), 2)
	if len(parts) < 2 {
		return "", nil, fmt.Errorf("invalid protocol:data format")
	}

	return string(parts[0]), parts[1], nil
}
