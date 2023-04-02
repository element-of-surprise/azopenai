package azopenai

import (
	"net/http"

	"github.com/element-of-surprise/azopenai/auth"
	"github.com/element-of-surprise/azopenai/clients/chat"
	"github.com/element-of-surprise/azopenai/clients/completions"
	"github.com/element-of-surprise/azopenai/clients/embeddings"
	"github.com/element-of-surprise/azopenai/rest"
)

// Client provides access to the Azure OpenAI service.
type Client struct {
	resourceName string
	deploymentID string

	auth   auth.Authorizer
	client *http.Client
	rest   *rest.Client
}

// Option provides optional arguments to the New constructor.
type Option func(*Client) error

// WithClient sets the HTTP client to use for requests.
func WithClient(c *http.Client) Option {
	return func(client *Client) error {
		client.client = c
		return nil
	}
}

// New creates a new instance of the Client.
func New(resourceName, deploymentID string, auth auth.Authorizer, options ...Option) (*Client, error) {
	c := &Client{
		resourceName: resourceName,
		deploymentID: deploymentID,
		auth:         auth,
	}

	for _, o := range options {
		if err := o(c); err != nil {
			return nil, err
		}
	}

	if c.client == nil {
		c.client = &http.Client{}
	}

	r, err := rest.New(resourceName, deploymentID, auth, rest.WithClient(c.client))
	if err != nil {
		return nil, err
	}
	c.rest = r

	return c, nil
}

// Completions will return a client for the Completions API. Completions attempt to return
// sentence completions give some input text. Each call returns a
// new instance of the client, not a shared instance.
func (c *Client) Completions() *completions.Client {
	return completions.New(c.rest)
}

// Embeddings will return a client for the Embeddings API. Embeddings converts text strings
// to vector representation that can be consumed by machine learning models. Each call returns a
// new instance of the client, not a shared instance.
func (c *Client) Embeddings() *embeddings.Client {
	return embeddings.New(c.rest)
}

// Chat will return a client for the Chat API. Chat provides a simple way to interact with
// the chat API for responding as a chat bot.
func (c *Client) Chat() *chat.Client {
	return chat.New(c.rest)
}
