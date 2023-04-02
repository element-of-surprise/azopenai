// Package rest provides access to the Azure OpenAI service via the REST API. This is
// a low-level package that provides access to the REST API directly. Most normal use
// cases will use the higher-level azopenai.Client and its sub-clients.
package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"text/template"

	"github.com/element-of-surprise/azopenai/auth"
	"github.com/element-of-surprise/azopenai/rest/messages/chat"
	"github.com/element-of-surprise/azopenai/rest/messages/completions"
	"github.com/element-of-surprise/azopenai/rest/messages/embeddings"
)

// APIVersion represents the version of the Azure OpenAI service this client is using.
const APIVersion = "2023-03-15-preview"

// Client provides access to the Azure OpenAI service via the REST API.
type Client struct {
	apiVersion   string
	resourceName string
	deploymentID string
	auth         auth.Authorizer
	client       *http.Client

	completionsURL *url.URL
	embeddingsURL  *url.URL
	chatURL        *url.URL
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

// New creates a new instance of the Client type.
func New(resourceName, deploymentID string, auth auth.Authorizer, options ...Option) (*Client, error) {
	var err error
	auth, err = auth.Validate()
	if err != nil {
		return nil, err
	}

	c := &Client{
		apiVersion:   APIVersion,
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

	if err := c.urls(); err != nil {
		return nil, err
	}

	return c, nil
}

// urls creates the URLs for the API endpoints based on the deployment ID. and API version. that
// was passed.
func (c *Client) urls() error {
	const (
		completions = "https://{{.resourceName}}.openai.azure.com/openai/deployments/{{.deploymentID}}/completions?api-version={{.apiVersion}}"
		embeddings  = "https://{{.resourceName}}.openai.azure.com/openai/deployments/{{.deploymentID}}/embeddings?api-version={{.apiVersion}}"
		chat        = "https://{{.resourceName}}.openai.azure.com/openai/deployments/{{.deploymentID}}/chat/completions?api-version={{.apiVersion}}"
	)

	type create struct {
		name   string
		urlStr string
		dest   *url.URL
	}

	l := []create{
		{"completions", completions, c.completionsURL},
		{"embeddings", embeddings, c.embeddingsURL},
		{"chat", chat, c.chatURL},
	}

	for _, v := range l {
		b := &strings.Builder{}
		t := template.Must(template.New(v.name).Parse(v.urlStr))
		if err := t.ExecuteTemplate(b, v.name, c); err != nil {
			return err
		}

		var err error
		v.dest, err = url.Parse(b.String())
		if err != nil {
			return err
		}
	}
	return nil
}

// requestsBuff is a pool of buffers used to marshal the request body.
var requestsBuff = newBufferPool()

// Complete sends a request to the Azure OpenAI service to complete the given prompt.
func (c *Client) Completions(ctx context.Context, req completions.Req) (completions.Resp, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return completions.Resp{}, err
	}
	resp, err := c.send(ctx, c.completionsURL, b)
	if err != nil {
		return completions.Resp{}, err
	}

	var msg completions.Resp
	if err := json.Unmarshal(resp, &msg); err != nil {
		return completions.Resp{}, fmt.Errorf("problem unmarshaling the response body: %w", err)
	}
	return msg, nil
}

// Embeddings sends a request to the Azure OpenAI service to get the embeddings for the given set of data.
func (c *Client) Embeddings(ctx context.Context, req embeddings.Req) (embeddings.Resp, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return embeddings.Resp{}, err
	}
	resp, err := c.send(ctx, c.completionsURL, b)
	if err != nil {
		return embeddings.Resp{}, err
	}

	var msg embeddings.Resp
	if err := json.Unmarshal(resp, &msg); err != nil {
		return embeddings.Resp{}, fmt.Errorf("problem unmarshaling the response body: %w", err)
	}

	sort.Slice(msg.Data, func(i, j int) bool {
		return msg.Data[i].Index < msg.Data[j].Index
	})

	return msg, nil
}

// Chat sends a request to the Azure OpenAI service to get responses to chat messages for the given set of data.
func (c *Client) Chat(ctx context.Context, req chat.Req) (chat.Resp, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return chat.Resp{}, err
	}
	resp, err := c.send(ctx, c.completionsURL, b)
	if err != nil {
		return chat.Resp{}, err
	}

	var msg chat.Resp
	if err := json.Unmarshal(resp, &msg); err != nil {
		return chat.Resp{}, fmt.Errorf("problem unmarshaling the response body: %w", err)
	}

	sort.Slice(msg.Choices, func(i, j int) bool {
		return msg.Choices[i].Index < msg.Choices[j].Index
	})

	return msg, nil
}

func (c *Client) send(ctx context.Context, addr *url.URL, msg []byte) ([]byte, error) {
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, "", nil)
	if err != nil {
		return nil, err
	}
	hreq.URL = addr

	if err := c.auth.Authorize(ctx, hreq); err != nil {
		return nil, err
	}

	buff := requestsBuff.Get()
	buff.Write(msg)
	hreq.Body = io.NopCloser(buff)
	defer func() {
		buff.Reset()
		requestsBuff.Put(buff)
	}()

	resp, err := c.client.Do(hreq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("problem reading the response body: %w", err)
	}

	return b, nil
}
