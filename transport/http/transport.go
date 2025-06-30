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
		"www.google.com",
		"api.cloudflare.com",
		"cdn.aws.com",
		"images.microsoft.com",
	}

	randomUserAgents = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		"AppleWebKit/537.36 (KHTML, like Gecko)",
		"Chrome/91.0.4472.124 Safari/537.36",
		"Edge/91.0.864.59",
	}

	randomPaths = []string{
		"/api/v1/users",
		"/graphql",
		"/rest/data",
		"/wp-json/wp/v2/posts",
	}
)

type HTTPCamouflage struct {
}

func (h *HTTPCamouflage) getRandomElement(list []string) string {
	return list[rand.Intn(len(list))]
}

func (h *HTTPCamouflage) Wrap(data []byte) []byte {
	host := h.getRandomElement(randomHosts)
	ua := h.getRandomElement(randomUserAgents)
	path := h.getRandomElement(randomPaths)

	request := fmt.Sprintf(
		"POST %s HTTP/1.1\r\n"+
			"Host: %s\r\n"+
			"User-Agent: %s\r\n"+
			"Content-Type: application/octet-stream\r\n"+
			"Content-Length: %d\r\n\r\n",
		path, host, ua, len(data))

	return append([]byte(request), data...)
}

func (h *HTTPCamouflage) Unwrap(data []byte) ([]byte, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	req, err := http.ReadRequest(reader)
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP request: %v", err)
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP request: %v", err)
	}
	return body, nil
}

func (h *HTTPCamouflage) CreateResponse(data []byte, protocol string) []byte {
	host := h.getRandomElement(randomHosts)
	ua := h.getRandomElement(randomUserAgents)

	// Calculate the total content length
	contentLength := len(protocol) + 1 + len(data) // protocol + ":" + data

	// Format the HTTP response
	response := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\n"+
			"Host: %s\r\n"+
			"User-Agent: %s\r\n"+
			"Content-Type: application/octet-stream\r\n"+
			"Content-Length: %d\r\n\r\n"+
			"%s:%s",
		host, ua,
		contentLength,
		protocol,
		data)

	return []byte(response)
}

func (h *HTTPCamouflage) DecodeResponse(resp []byte) (string, []byte, error) {
	reader := bufio.NewReader(bytes.NewReader(resp))
	response, err := http.ReadResponse(reader, nil)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse HTTP response: %v", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Split protocol and data
	parts := bytes.SplitN(body, []byte(":"), 2)
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("invalid response format, missing protocol separator")
	}

	return string(parts[0]), parts[1], nil
}
