package orcclient

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (c *orcClient) Get(url string) (*http.Response, error) {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

func (c *orcClient) Head(url string) (*http.Response, error) {
	r, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

func (c *orcClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	r, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", contentType)
	return c.Do(r)
}

func (c *orcClient) PostForm(url string, data url.Values) (*http.Response, error) {
	body := data.Encode()
	bodyReader := strings.NewReader(body)
	contentType := "application/x-www-form-urlencoded"
	return c.Post(url, contentType, bodyReader)
}
