package plural

import (
	"context"
	"fmt"
	"github.com/michaeljguarino/graphql"
	"os"
)

type Config struct {
	Token    string
	Endpoint string
}

type Client struct {
	gqlClient *graphql.Client
	config    *Config
}

func NewConfig() *Config {
	return &Config{
		Token:    os.Getenv("PLURAL_ACCESS_TOKEN"),
		Endpoint: os.Getenv("PLURAL_ENDPOINT"),
	}
}

func NewClient(conf *Config) *Client {
	base := conf.BaseUrl()
	return &Client{graphql.NewClient(base + "/gql"), conf}
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

func (client *Client) Build(doc string) *graphql.Request {
	req := graphql.NewRequest(doc)
	req.Header.Set("Authorization", "Bearer "+client.config.Token)
	return req
}

func (client *Client) Run(req *graphql.Request, resp interface{}) error {
	return client.gqlClient.Run(context.Background(), req, &resp)
}
