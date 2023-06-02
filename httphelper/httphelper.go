package httphelper

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func Request(client http.Client, method, url string, body io.Reader, token string, counter int) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)

	if token != "" {
		req.Header.Set("x-auth-token", token)
	}
	for i := 0; i < counter; i++ {

		resp, err2 := client.Do(req)

		if err != nil {
			return nil, fmt.Errorf("cannot create request: %w", err)
		}
		if err2 != nil {
			return nil, fmt.Errorf("cannot create request: %w", err2)
		}

		if resp.StatusCode == http.StatusOK {
			return resp, nil
		}
		time.Sleep(300 * time.Millisecond)
	}
	return nil, fmt.Errorf("cannot access server")
}
