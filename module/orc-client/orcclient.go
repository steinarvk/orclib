package orcclient

import (
	"io"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
)

type Client interface {
	Do(req *http.Request) (*http.Response, error)
	// CloseIdleConnections()
	Get(url string) (*http.Response, error)
	Head(url string) (*http.Response, error)
	Post(url, contentType string, body io.Reader) (*http.Response, error)
	PostForm(url string, data url.Values) (*http.Response, error)
}

type orcClient struct {
	clientName string
	m          *Module
	httpClient Client
}

func (m *Module) New(name string) (Client, error) {
	underlyingClient, err := m.createClient()
	if err != nil {
		return nil, err
	}
	return &orcClient{
		clientName: name,
		httpClient: underlyingClient,
		m:          m,
	}, nil
}

func (m *Module) MustNew(name string) Client {
	rv, err := m.New(name)
	if err != nil {
		logrus.Fatalf("Unable to create client %q", name)
	}
	return rv
}
