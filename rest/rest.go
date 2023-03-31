// Package rest provides access to the Azure OpenAI service via the REST API. This is
// a low-level package that provides access to the REST API directly.
package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"text/template"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"

	"github.com/element-of-surprise/azopenai/rest/messages"
)

const apiVersion = "2023-03-15-preview"

// Client provides access to the Azure OpenAI service via the REST API.
type Client struct {
	apiVersion   string
	resourceName string
	deploymentID string
	auth         Auth
	client       *http.Client

	completionsURL *url.URL
}

// Auth provides authorization options for authenticating to the Azure service.
type Auth struct {
	// AzIdentity provides authentication/authorization using the AzIdentity package.
	AzIdentity AzIdentity
}

func (a Auth) validate() error {
	if err := a.AzIdentity.validate(); err != nil {
		return err
	}
	return nil
}

// AzIdentity provides authentication/authorization using the AzIdentity package.
type AzIdentity struct {
	// Credential is the credential used to authenticate to the service.
	// This can be acquired by using one of the methods in:
	// https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity
	Credential azcore.TokenCredential
	// Policy provides scopes for the token request.
	Policy policy.TokenRequestOptions
}

func (a AzIdentity) validate() error {
	if a.Credential == nil {
		return fmt.Errorf("missing Credential")
	}
	return nil
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
func New(resourceName, deploymentID string, auth Auth, options ...Option) (*Client, error) {
	if err := auth.validate(); err != nil {
		return nil, err
	}

	c := &Client{
		apiVersion:   apiVersion,
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
	const completions = "https://{{.resourceName}}.openai.azure.com/openai/deployments/{{.deploymentID}}/completions?api-version={{.apiVersion}}"

	b := &strings.Builder{}
	t := template.Must(template.New("completions").Parse(completions))
	if err := t.ExecuteTemplate(b, "completions", c); err != nil {
		return err
	}

	var err error
	c.completionsURL, err = url.Parse(b.String())
	if err != nil {
		return err
	}
	return nil
}

// authorize adds the authorization header to the request.
func (c *Client) authorize(ctx context.Context, req *http.Request) error {
	t, err := c.auth.AzIdentity.Credential.GetToken(ctx, c.auth.AzIdentity.Policy)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", t))
	return err
}

// requestsBuff is a pool of buffers used to marshal the request body.
var requestsBuff = &sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

func (c *Client) Completions(ctx context.Context, req messages.PromptRequest) (messages.PromptResponse, error) {
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, "", nil)
	if err != nil {
		return messages.PromptResponse{}, err
	}
	hreq.URL = c.completionsURL

	if err := c.authorize(ctx, hreq); err != nil {
		return messages.PromptResponse{}, err
	}

	b, err := json.Marshal(req)
	if err != nil {
		return messages.PromptResponse{}, err
	}
	buff := requestsBuff.Get().(*bytes.Buffer)
	buff.Write(b)
	hreq.Body = io.NopCloser(buff)
	defer func() {
		buff.Reset()
		requestsBuff.Put(buff)
	}()

	resp, err := c.client.Do(hreq)
	if err != nil {
		return messages.PromptResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return messages.PromptResponse{}, fmt.Errorf("status code %d", resp.StatusCode)
	}

	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return messages.PromptResponse{}, fmt.Errorf("problem reading the response body: %w", err)
	}

	var res messages.PromptResponse
	if err := json.Unmarshal(b, &res); err != nil {
		return messages.PromptResponse{}, fmt.Errorf("problem unmarshaling the response body: %w", err)
	}
	return res, nil
}
