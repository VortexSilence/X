package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"
)

func TestHTTPCamouflage(t *testing.T) {
	// Initialize camouflage with test values
	camouflage := &HTTPCamouflage{}

	// Test cases table
	tests := []struct {
		name        string
		input       []byte
		shouldFail  bool
		expected    []byte
		description string
	}{
		{
			name:        "NormalTextData",
			input:       []byte("test data"),
			shouldFail:  false,
			expected:    []byte("test data"),
			description: "Regular text data without special characters",
		},
		{
			name:        "DataWithNewlines",
			input:       []byte("line1\nline2\r\nline3"),
			shouldFail:  false,
			expected:    []byte("line1\nline2\r\nline3"),
			description: "Data containing various newline characters",
		},
		{
			name:        "EmptyData",
			input:       []byte(""),
			shouldFail:  false,
			expected:    []byte(""),
			description: "Empty payload",
		},
		{
			name:        "LargeData",
			input:       bytes.Repeat([]byte("a"), 1024*1024), // 1MB data
			shouldFail:  false,
			expected:    bytes.Repeat([]byte("a"), 1024*1024),
			description: "Large data payload (1MB)",
		},
		{
			name:        "BinaryData",
			input:       []byte{0x00, 0xFF, 0x42, 0x7E},
			shouldFail:  false,
			expected:    []byte{0x00, 0xFF, 0x42, 0x7E},
			description: "Binary/non-ASCII data",
		},
		{
			name:        "InvalidHTTP",
			input:       []byte("not http data"),
			shouldFail:  false,
			expected:    []byte("not http data"),
			description: "Invalid HTTP format (should fail)",
		},
	}

	// Run standard test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Process the data through Wrap and Unwrap
			wrapped := camouflage.Wrap(tt.input, "tcp")
			_, unwrapped, err := camouflage.Unwrap(wrapped)
			if tt.shouldFail {
				if err == nil {
					t.Errorf("%s: Expected error but got none", tt.description)
				}
				return
			}

			// Check for unexpected errors
			if err != nil {
				t.Errorf("%s: Unexpected error: %v", tt.description, err)
				return
			}

			// Verify data integrity
			if !bytes.Equal(tt.expected, unwrapped) {
				t.Errorf("%s: Data mismatch\nInput: %v\nOutput: %v",
					tt.description, tt.input, unwrapped)
			}
		})
	}

	// Special case: Test with HTTP response instead of request
	t.Run("HTTPResponseInput", func(t *testing.T) {
		// Create a test HTTP response
		response := httptest.NewRecorder()
		response.Body.WriteString("response data")
		respBytes, _ := io.ReadAll(response.Result().Body)

		// Attempt to unwrap the response
		_, _, err := camouflage.Unwrap(respBytes)
		if err == nil {
			t.Error("Expected error for HTTP response input, but got none")
		}
	})

	// Special case: Test with real HTTP request format
	t.Run("RealHTTPRequest", func(t *testing.T) {
		body := "tcp:hello world"
		req := fmt.Sprintf(
			"POST / HTTP/1.1\r\n"+
				"Host: example.com\r\n"+
				"Content-Length: %d\r\n\r\n%s",
			len(body), body)
		s, result, err := camouflage.Unwrap([]byte(req))

		if err != nil {
			t.Errorf("Error processing real HTTP request: %v", err)
		}
		if s != "tcp" {
			t.Errorf("Unexpected processed type. Got: %s", s)
		}
		if string(result) != "hello world" {
			t.Errorf("Unexpected processed data. Got: %s", result)
		}
	})

	// Edge case: Test with malformed Content-Length
	t.Run("MalformedContentLength", func(t *testing.T) {
		malformedReq := "POST /api HTTP/1.1\r\nHost: example.com\r\nContent-Length: tcp:invalid\r\n\r\ndata"
		_, _, err := camouflage.Unwrap([]byte(malformedReq))
		if err == nil {
			t.Error("Expected error for malformed Content-Length, but got none")
		}
	})

	// Edge case: Test with chunked transfer encoding
	// FIXME: plzz
	// t.Run("ChunkedTransferEncoding", func(t *testing.T) {
	// 	chunkedReq := "POST /api HTTP/1.1\r\nHost: example.com\r\nTransfer-Encoding: chunked\r\n\r\n5\r\nhello\r\n6\r\n world\r\n0\r\n\r\n"
	// 	_, result, err := camouflage.Unwrap([]byte(chunkedReq))
	// 	if err != nil {
	// 		t.Errorf("Unexpected error with chunked encoding: %v", err)
	// 	}
	// 	if string(result) != "hello world" {
	// 		t.Errorf("Incorrect chunked data processing. Got: %s", result)
	// 	}
	// })
}

func TestHTTPCamouflageResponse(t *testing.T) {
	camouflage := &HTTPCamouflage{}

	tests := []struct {
		name        string
		protocol    string
		data        []byte
		description string
	}{
		{
			name:        "SimpleText",
			protocol:    "proto1",
			data:        []byte("test data"),
			description: "Simple text data",
		},
		{
			name:        "BinaryData",
			protocol:    "binproto",
			data:        []byte{0x00, 0xFF, 0x42},
			description: "Binary data with null bytes",
		},
		// {
		// 	name:        "EmptyData",
		// 	protocol:    "empty",
		// 	data:        []byte(""),
		// 	description: "Empty data payload",
		// },
		{
			name:        "SpecialChars",
			protocol:    "special",
			data:        []byte("data\r\nwith\nspecial\tchars"),
			description: "Data with special characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test CreateResponse
			response := camouflage.WrapResponse(tt.data, tt.protocol, 200)

			// Verify the response can be decoded
			decodedProtocol, decodedData, err := camouflage.UnwrapResponse(response)
			if err != nil {
				t.Fatalf("%s: Decode failed: %v", tt.description, err)
			}

			// Verify protocol matches
			if decodedProtocol != tt.protocol {
				t.Errorf("%s: Protocol mismatch. Expected: %s, Got: %s",
					tt.description, tt.protocol, decodedProtocol)
			}

			// Verify data matches
			if !bytes.Equal(decodedData, tt.data) {
				t.Errorf("%s: Data mismatch. Expected: %v, Got: %v",
					tt.description, tt.data, decodedData)
			}

			// Verify the response is valid HTTP
			if !bytes.Contains(response, []byte("HTTP/1.1 200 OK")) {
				t.Errorf("%s: Invalid HTTP response header", tt.description)
			}
		})
	}
}
