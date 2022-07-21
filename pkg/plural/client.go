package plural

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/pluralsh/gqlclient"
)

type authedTransport struct {
	key     string
	wrapped http.RoundTripper
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.key)
	return t.wrapped.RoundTrip(req)
}

type Config struct {
	Token    string
	Endpoint string
}

type Client struct {
	ctx          context.Context
	pluralClient *gqlclient.Client
	config       *Config
}

func NewConfig() *Config {
	return &Config{
		Token:    os.Getenv("PLURAL_ACCESS_TOKEN"),
		Endpoint: os.Getenv("PLURAL_ENDPOINT"),
	}
}

func NewClient(conf *Config) *Client {
	base := conf.BaseUrl()
	httpClient := http.Client{
		Transport: &authedTransport{
			key:     conf.Token,
			wrapped: http.DefaultTransport,
		},
	}
	endpoint := base + "/gql"
	return &Client{
		ctx:          context.Background(),
		pluralClient: gqlclient.NewClient(&httpClient, endpoint),
		config:       conf,
	}
}

func (c *Config) BaseUrl() string {
	return fmt.Sprintf("https://%s", c.PluralEndpoint())
}

func (c *Config) PluralEndpoint() string {
	if c.Endpoint == "" {
		return "app.plural.sh"
	}

	return c.Endpoint
}
