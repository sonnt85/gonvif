package networking

import (
	"bytes"
	"net/http"
	"time"
)

// SendSoap send soap message
func SendSoap(httpClient *http.Client, endpoint string, message string, timeout ...time.Duration) (*http.Response, error) {
	if len(timeout) != 0 {
		httpClient.Timeout = timeout[0]
	}
	resp, err := httpClient.Post(endpoint, "application/soap+xml; charset=utf-8", bytes.NewBufferString(message))
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// SendSoapWithTimeout send soap message with timeOut
func SendSoapWithTimeout(httpClient *http.Client, endpoint string, message []byte, timeout time.Duration) (*http.Response, error) {
	httpClient.Timeout = timeout
	return httpClient.Post(endpoint, "application/soap+xml; charset=utf-8", bytes.NewReader(message))
}
