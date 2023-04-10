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
	"strings"
	"sync"
	"text/template"

	"github.com/element-of-surprise/azopenai/auth"
	"github.com/element-of-surprise/azopenai/errors"
	"github.com/element-of-surprise/azopenai/rest/messages/chat"
	"github.com/element-of-surprise/azopenai/rest/messages/completions"
	"github.com/element-of-surprise/azopenai/rest/messages/embeddings"
)

// APIVersion represents the version of the Azure OpenAI service this client is using.
const APIVersion = "2023-03-15-preview"

type templVars struct {
	ResourceName string
	DeploymentID string
	APIVersion   string
}

type deployments map[string]*url.URL

type endpoints struct {
	temps *template.Template
	mu    sync.Mutex // Protects m
	m     map[endpointType]deployments
}

type endpointType string

const (
	unknownTmpl     endpointType = ""
	completionsTmpl endpointType = "completions"
	embeddingsTmpl  endpointType = "embeddings"
	chatTmpl        endpointType = "chat"
)

func newEndpoints() *endpoints {
	const (
		completions = "https://{{.ResourceName}}.openai.azure.com/openai/deployments/{{.DeploymentID}}/completions?api-version={{.APIVersion}}"
		embeddings  = "https://{{.ResourceName}}.openai.azure.com/openai/deployments/{{.DeploymentID}}/embeddings?api-version={{.APIVersion}}"
		chat        = "https://{{.ResourceName}}.openai.azure.com/openai/deployments/{{.DeploymentID}}/chat/completions?api-version={{.APIVersion}}"
	)

	temps := &template.Template{}
	temps = template.Must(temps.New(string(completionsTmpl)).Parse(completions))
	temps = template.Must(temps.New(string(embeddingsTmpl)).Parse(embeddings))
	temps = template.Must(temps.New(string(chatTmpl)).Parse(chat))

	return &endpoints{
		temps: temps,
		m:     make(map[endpointType]deployments),
	}
}

func (e *endpoints) url(eType endpointType, deploymentID string, vars templVars) (*url.URL, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	deploy := e.m[eType]
	if deploy == nil {
		deploy = deployments{}
		e.m[eType] = deploy
	}

	u := deploy[deploymentID]
	if u == nil {
		vars.DeploymentID = deploymentID
		u, err := e.set(eType, vars)
		if err != nil {
			return nil, err
		}
		e.m[eType][deploymentID] = u
		return u, nil
	}
	return u, nil
}

func (e *endpoints) set(et endpointType, vars templVars) (*url.URL, error) {
	b := &strings.Builder{}
	if err := e.temps.ExecuteTemplate(b, string(et), vars); err != nil {
		return nil, err
	}

	return url.Parse(b.String())
}

// Client provides access to the Azure OpenAI service via the REST API.
type Client struct {
	auth   auth.Authorizer
	client *http.Client

	vars templVars

	completionsURL *url.URL
	embeddingsURL  *url.URL
	chatURL        *url.URL

	endpoints *endpoints
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
func New(resourceName string, auth auth.Authorizer, options ...Option) (*Client, error) {
	var err error
	auth, err = auth.Validate()
	if err != nil {
		return nil, err
	}

	c := &Client{
		vars: templVars{
			ResourceName: resourceName,
			APIVersion:   APIVersion,
		},
		endpoints: newEndpoints(),
		auth:      auth,
	}
	for _, o := range options {
		if err := o(c); err != nil {
			return nil, err
		}
	}

	if c.client == nil {
		c.client = &http.Client{}
	}

	return c, nil
}

// requestsBuff is a pool of buffers used to marshal the request body.
var requestsBuff = newBufferPool()

// Complete sends a request to the Azure OpenAI service to complete the given prompt.
func (c *Client) Completions(ctx context.Context, deploymentID string, req completions.Req) (completions.Resp, error) {
	u, err := c.endpoints.url(completionsTmpl, deploymentID, c.vars)
	if err != nil {
		return completions.Resp{}, err
	}

	b, err := json.Marshal(req)
	if err != nil {
		return completions.Resp{}, err
	}
	resp, err := c.send(ctx, u, b)
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
func (c *Client) CompletionsStream(ctx context.Context, deploymentID string, req completions.Req) chan StreamRecv[completions.Resp] {
	ch := make(chan StreamRecv[completions.Resp], 1)

	u, err := c.endpoints.url(completionsTmpl, deploymentID, c.vars)
	if err != nil {
		ch <- StreamRecv[completions.Resp]{Err: err}
		return ch
	}

	req.Stream = true
	b, err := json.Marshal(req)
	if err != nil {
		ch <- StreamRecv[completions.Resp]{Err: err}
		return ch
	}

	go func() {
		defer close(ch)

		responses, err := c.stream(ctx, u, b)
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
func (c *Client) Embeddings(ctx context.Context, deploymentID string, req embeddings.Req) (embeddings.Resp, error) {
	u, err := c.endpoints.url(embeddingsTmpl, deploymentID, c.vars)
	if err != nil {
		return embeddings.Resp{}, err
	}

	b, err := json.Marshal(req)
	if err != nil {
		return embeddings.Resp{}, err
	}
	resp, err := c.send(ctx, u, b)
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
func (c *Client) Chat(ctx context.Context, deploymentID string, req chat.Req) (chat.Resp, error) {
	u, err := c.endpoints.url(chatTmpl, deploymentID, c.vars)
	if err != nil {
		return chat.Resp{}, err
	}

	b, err := json.Marshal(req)
	if err != nil {
		return chat.Resp{}, err
	}
	resp, err := c.send(ctx, u, b)
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
		return nil, specErr(resp)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("problem reading the response body: %w", err)
	}

	return b, nil
}

var bufIOs = sync.Pool{
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
		return nil, specErr(resp)
	}

	ch := make(chan StreamRecv[[]byte], 1)
	go func() {
		defer close(ch)

		bio := bufIOs.Get().(*bufio.Reader)
		bio.Reset(resp.Body)
		defer bufIOs.Put(bio)

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

func specErr(resp *http.Response) error {
	msg, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.StatusCode{
			Message:    string(msg),
			StatusCode: resp.StatusCode,
		}
	}

	m := map[string]any{}
	if err := json.Unmarshal(msg, &m); err != nil {
		return errors.StatusCode{
			Message:    string(msg),
			StatusCode: resp.StatusCode,
		}
	}
	return errors.JSON{
		Message:    string(msg),
		JSON:       m,
		StatusCode: resp.StatusCode,
	}
}

// StreamRecv is used to receive data from a stream.
type StreamRecv[T any] struct {
	// Data is data sent by the stream.
	Data T
	// Err is an error related to the stream. The stream is terminated after this.
	Err error
}
