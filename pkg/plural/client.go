package plural

import (
	"context"
	"github.com/michaeljguarino/graphql"
)

type Config struct {
	Token    string
	Endpoint string
}

type Client struct {
	gqlClient *graphql.Client
	config    *Config
}

func NewClient(conf *Config) *Client {
	base := conf.BaseUrl()
	return &Client{graphql.NewClient(base + "/gql"), conf}
}

func (c *Config) BaseUrl() string {
	host := "https://app.plural.sh"
	if c.Endpoint != "" {
		host = "https://" + c.Endpoint
	}
	return host
}

func (client *Client) Build(doc string) *graphql.Request {
	req := graphql.NewRequest(doc)
	req.Header.Set("Authorization", "Bearer "+client.config.Token)
	return req
}

func (client *Client) Run(req *graphql.Request, resp interface{}) error {
	return client.gqlClient.Run(context.Background(), req, &resp)
}
