// Package rest provides access to the Azure OpenAI service via the REST API. This is
// a low-level package that provides access to the REST API directly. Most normal use
// cases will use the higher-level azopenai.Client and its sub-clients.
package rest

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"sync"

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
		completions = "https://%s.openai.azure.com/openai/deployments/%s/completions?api-version=%s"
		embeddings  = "https://%s.openai.azure.com/openai/deployments/%s/embeddings?api-version=%s"
		chat        = "https://%s.openai.azure.com/openai/deployments/%s/chat/completions?api-version=%s"
	)

	var err error
	c.completionsURL, err = url.Parse(fmt.Sprintf(completions, c.resourceName, c.deploymentID, c.apiVersion))
	if err != nil {
		return err
	}
	c.embeddingsURL, err = url.Parse(fmt.Sprintf(embeddings, c.resourceName, c.deploymentID, c.apiVersion))
	if err != nil {
		return err
	}
	c.chatURL, err = url.Parse(fmt.Sprintf(chat, c.resourceName, c.deploymentID, c.apiVersion))
	if err != nil {
		return err
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

// CompletionsStream is the same as Completions, except that as the service accumulates tokens to respond
// to the request, it will stream the results back to the client. The client can stop the stream by cancelling
// the context.
func (c *Client) CompletionsStream(ctx context.Context, req completions.Req) chan StreamRecv[completions.Resp] {
	ch := make(chan StreamRecv[completions.Resp], 1)
	req.Stream = true
	b, err := json.Marshal(req)
	if err != nil {
		ch <- StreamRecv[completions.Resp]{Err: err}
		return ch
	}

	go func() {
		defer close(ch)

		responses, err := c.stream(ctx, c.completionsURL, b)
		if err != nil {
			ch <- StreamRecv[completions.Resp]{Err: err}
			return
		}

		for response := range responses {
			var msg completions.Resp
			if err := json.Unmarshal(response.Data, &msg); err != nil {
				ch <- StreamRecv[completions.Resp]{Err: fmt.Errorf("problem unmarshaling the response body: %w", err)}
				return
			}
			ch <- StreamRecv[completions.Resp]{Data: msg}
		}
	}()

	return ch
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

	resp, err := c.send(ctx, c.chatURL, b)
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
	hreq.Host = addr.Host
	hreq.URL = addr

	if err := c.auth.Authorize(ctx, hreq); err != nil {
		return nil, err
	}

	buff := requestsBuff.Get()
	defer requestsBuff.Put(buff)

	buff.Reset(msg)
	hreq.Body = buff

	resp, err := c.client.Do(hreq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("status code %d and error reading result body", resp.StatusCode)
		}
		return nil, fmt.Errorf("status code %d and result body %q", resp.StatusCode, string(b))
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("problem reading the response body: %w", err)
	}

	return b, nil
}

var bios = sync.Pool{
	New: func() any {
		return bufio.NewReader(nil)
	},
}

var streamDone = []byte("[DONE]")
var streamHeader = []byte("data: ")

func (c *Client) stream(ctx context.Context, addr *url.URL, msg []byte) (chan StreamRecv[[]byte], error) {
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, "", nil)
	if err != nil {
		return nil, err
	}
	hreq.URL = addr

	if err := c.auth.Authorize(ctx, hreq); err != nil {
		return nil, err
	}

	buff := requestsBuff.Get()
	defer requestsBuff.Put(buff)

	buff.Reset(msg)
	hreq.Body = buff

	resp, err := c.client.Do(hreq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	ch := make(chan StreamRecv[[]byte], 1)
	go func() {
		defer close(ch)

		bio := bios.Get().(*bufio.Reader)
		bio.Reset(resp.Body)
		defer bios.Put(bio)

		for {
			line, err := bio.ReadBytes('\n')
			if err != nil {
				ch <- StreamRecv[[]byte]{Err: err}
				return
			}
			line = bytes.TrimSpace(line)

			if !bytes.HasPrefix(line, streamHeader) {
				// This indicates an empty message. We may want to put a limit on the number of empty messages.
				// For now, we just ignore them.
				continue
			}
			line = bytes.TrimPrefix(line, streamHeader)

			// This indicates the end of the stream.
			if bytes.Equal(line, streamDone) {
				return
			}

			ch <- StreamRecv[[]byte]{Data: line}
		}
	}()

	return ch, nil
}

// StreamRecv is used to receive data from a stream.
type StreamRecv[T any] struct {
	// Data is data sent by the stream.
	Data T
	// Err is an error related to the stream. The stream is terminated after this.
	Err error
}
