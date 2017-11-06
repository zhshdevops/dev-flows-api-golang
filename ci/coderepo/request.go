package coderepo

import (
	"net/http"
	"time"
	"io"
)

func HttpClientRequest(method, url string, body io.Reader, header map[string]string) (*http.Response, error) {

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	if len(header) != 0 {
		for key, value := range header {
			req.Header.Add(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	resp.Header.Set("Content-Type", "application/json; charset=utf-8")

	return resp, err
}
