/*
Package azopenai provides a client for the Azure OpenAI Service.

This package allows access to Azure OpenAI Service using either an API key or
using [AzIdentity] to authenticate with Azure Active Directory.

The client is split into three sub-clients: completions, chat, and embeddings.
Each of these sub-clients provides access to the corresponding API endpoints. You
can access each of these sub-clients by calling the corresponding method on the main client.
They will all share the same authentication and http.Client.

Creating a Client with an API Key:

	client, err := azopenai.New(resourceName, auth.Authorizer{ApiKey: apiKey})
	if err != nil {
		return err
	}

Creating a Client with AzIdentity and default Azure credentials:

	client, err := New(resourceName, auth.Authorizer{AzIdentity: azidentity.NewDefaultAzureCredential()})
	if err != nil {
		return err
	}

Creating a Client with AzIdentity and a system [Managed Identity for Azure Resources] credential:

	client, err := New(resourceName, auth.Authorizer{AzIdentity: azidentity.NewMSICredential()})
	if err != nil {
		return err
	}

Creating a Client with AzIdentity and a user [Managed Identity for Azure Resources] credential:

	client, err := New(resourceName, auth.Authorizer{AzIdentity: azidentity.NewMSICredential("yourmsiid")})
	if err != nil {
		return err
	}

It should be noted that the New() method will not return an error if your credentials
are invalid. Only after calling a method on the sub-clients will you get an error if your
credentials or resource/deployment names are invalid.

If your program needs to terminate or deal with a chat client issue early in its runtime,
it is suggested to make a call to one of the APIs to ensure that your credentials are valid after
creating the client.

[AzIdentity]: https://github.com/Azure/azure-sdk-for-go/blob/main/sdk/azidentity/README.md
[Managed Identity for Azure Resources]: https://learn.microsoft.com/azure/active-directory/managed-identities-azure-resources/overview
*/
package azopenai

import (
	"net/http"

	"github.com/element-of-surprise/azopenai/auth"
	"github.com/element-of-surprise/azopenai/clients/chat"
	"github.com/element-of-surprise/azopenai/clients/completions"
	"github.com/element-of-surprise/azopenai/clients/embeddings"
	"github.com/element-of-surprise/azopenai/rest"
)

// Client provides access to the Azure OpenAI Service.
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
func New(resourceName string, auth auth.Authorizer, options ...Option) (*Client, error) {
	c := &Client{
		resourceName: resourceName,
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

	r, err := rest.New(resourceName, auth, rest.WithClient(c.client))
	if err != nil {
		return nil, err
	}
	c.rest = r

	return c, nil
}

// Completions will return a client for the Completions API. Completions attempt to return
// sentence completions give some input text. Each call returns a
// new instance of the client, not a shared instance.
func (c *Client) Completions(deploymentID string) *completions.Client {
	return completions.New(deploymentID, c.rest)
}

// Embeddings will return a client for the Embeddings API. Embeddings converts text strings
// to vector representation that can be consumed by machine learning models. Each call returns a
// new instance of the client, not a shared instance.
func (c *Client) Embeddings(deploymentID string) *embeddings.Client {
	return embeddings.New(deploymentID, c.rest)
}

// Chat will return a client for the Chat API. Chat provides a simple way to interact with
// the chat API for responding as a chat bot.
func (c *Client) Chat(deploymentID string) *chat.Client {
	return chat.New(deploymentID, c.rest)
}
