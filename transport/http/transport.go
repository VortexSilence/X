package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type HTTPCamouflage struct {
	Host      string
	Path      string
	UserAgent string
}

func (h *HTTPCamouflage) Wrap(data []byte) []byte {
	request := fmt.Sprintf(
		"POST %s HTTP/1.1\r\n"+
			"Host: %s\r\n"+
			"User-Agent: %s\r\n"+
			"Content-Type: application/octet-stream\r\n"+
			"Content-Length: %d\r\n\r\n",
		h.Path, h.Host, h.UserAgent, len(data))

	return append([]byte(request), data...)
}

func (h *HTTPCamouflage) Unwrap(data []byte) ([]byte, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	req, err := http.ReadRequest(reader)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
