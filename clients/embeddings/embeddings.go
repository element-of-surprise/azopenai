/*
Package embeddings provides access to the Embeddings API. Embeddings allows you to generate
vector representations of text for the purpose of model training.

The simplest way to create a Client is by using the azopenai.Client.Embeddings() method.

Using this API is simple:

	embeddingsClient := client.Embeddings("deploymentID")
	ctx := context.Background()
	text := []string{"The food was delicious and the waiter..."}
	resp, err := embeddingsClient.Call(ctx, text)
	if err != nil {
		return err
	}
	fmt.Printf("%v", resp.Results)

You can also set the default parameters for the client:

	embeddingsClient := client.Embeddings("deploymentID")

	// This creates a new instance of CallParams with the default values.
	// We then modify then and set them on the client. They will be used on
	// every call unless you override them on a specific call.
	params := embeddings.CallParams{}
	params.User = "element-of-surprise"
	embeddingsClient.SetParams(params)

	text := []string{"The food was delicious and the waiter..."}
	resp, err := embeddingsClient.Call(context.Background(), text)
	if err != nil {
		return err
	}
	fmt.Printf("%v", resp.Results)

You can also override the parameters on a per-call basis:

	text := []string{"The food was delicious and the waiter..."}
	resp, err := embeddingsClient.Call(context.Background(), text, embeddings.WithCallParams(params))
	if err != nil {
		return err
	}
	fmt.Printf("%v", resp.Results)
*/
package embeddings

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/element-of-surprise/azopenai/rest"
	"github.com/element-of-surprise/azopenai/rest/messages/embeddings"
)

// Client provides access to the Embeddings API. Embeddings allows converting text strings
// to vector representation that can be consumed by machine learning models.
type Client struct {
	deploymentID string
	rest         *rest.Client

	CallParams atomic.Pointer[CallParams]
}

// New creates a new instance of the Client type from the rest.Client. This is generally
// not used directly, but is used by the azopenai.Client.
func New(deploymentID string, rest *rest.Client) *Client {
	return &Client{
		deploymentID: deploymentID,
		rest:         rest,
	}
}

// CallParams are the parameters used on each call to the embeddings service. These
// are all optional fields. You can set this on the client and override it on a per-call
// basis.
type CallParams struct {
	// User is a unique identifier representing your end-user, which can help monitoring and detecting abuse.
	User string `json:"user,omitempty"`
	// Type is the embedding search to use. This is optional.
	Type string `json:"input_type,omitempty"`
	// Model is the model ID to use. This is optional.
	Model string `json:"model,omitempty"`
}

func (c CallParams) toEmbeddingsRequest() embeddings.Req {
	return embeddings.Req{
		User:  c.User,
		Type:  c.Type,
		Model: c.Model,
	}
}

// SetParams sets the CallParams for the client. This will be used for all calls unless
// overridden by a CallOption.
func (c *Client) SetParams(params CallParams) {
	c.CallParams.Store(&params)
}

// Embeddings returns the embeddings for the given set of text.
type Embeddings struct {
	// Results is a set of embeddings([]float64), one for each input sent.
	Results [][]float64

	// RestReq is the raw REST request sent to the server. This is only set if requested
	// with a CallOption.
	RestReq embeddings.Req
	// RestResp is the raw REST response from the server. This is only set if requested
	// with a CallOption.
	RestResp embeddings.Resp
}

type callOptions struct {
	CallParams    CallParams
	DeploymentID  string
	setCallParams bool

	RestReq        bool
	RestResp       bool
	RemoveNewlines bool
}

// CallOption is an optional argument for the Call method.
type CallOption func(options *callOptions) error

// WithCallParams sets the CallParams for the call. If not set, the call params set for
// the client will be used. If those weren't set, the default call options are used.
func WithCallParams(params CallParams) CallOption {
	return func(o *callOptions) error {
		o.CallParams = params
		o.setCallParams = true
		return nil
	}
}

// WithDeploymentID sets the deployment ID to use for the call. If not set, the deploymentID
// set on the client will be used.
func WithDeploymentID(deploymentID string) CallOption {
	return func(o *callOptions) error {
		o.DeploymentID = deploymentID
		return nil
	}
}

// WithRest sets whether to return the raw REST request and response. This is useful for
// debugging purposes.
func WithRest(req, resp bool) CallOption {
	return func(o *callOptions) error {
		o.RestReq = req
		o.RestResp = resp
		return nil
	}
}

// WithNewlineRemoval sets whether to remove newlines from the response and change to
// a space. This is useful when creating embeddings for text that doesn't represent
// programming code, as it has been observed that newlines will cause less optimal results.
func WithNewlineRemoval() CallOption {
	return func(o *callOptions) error {
		o.RemoveNewlines = true
		return nil
	}
}

// Call makes a call to the Embeddings API endpoint and returns the embeddings for the tokens.
func (c *Client) Call(ctx context.Context, text []string, options ...CallOption) (Embeddings, error) {
	callOptions := callOptions{}
	for _, o := range options {
		if err := o(&callOptions); err != nil {
			return Embeddings{}, err
		}
	}
	if !callOptions.setCallParams {
		callOptions.CallParams = CallParams{}
		p := c.CallParams.Load()
		if p != nil {
			callOptions.CallParams = *p
		}
	}

	// Remove newlines if requested.
	if callOptions.RemoveNewlines {
		for i := 0; i < len(text); i++ {
			text[i] = strings.ReplaceAll(text[i], "\n", " ")
		}
	}

	req := callOptions.CallParams.toEmbeddingsRequest()
	req.Input = text

	deploymentID := c.deploymentID
	if callOptions.DeploymentID != "" {
		deploymentID = callOptions.DeploymentID
	}

	resp, err := c.rest.Embeddings(ctx, deploymentID, req)
	if err != nil {
		return Embeddings{}, err
	}

	emb := Embeddings{Results: make([][]float64, len(resp.Data))}
	for i, data := range resp.Data {
		r := emb.Results[i]
		r = append(r, data.Embedding...)
		emb.Results[i] = r
	}

	if callOptions.RestReq {
		emb.RestReq = req
	}
	if callOptions.RestResp {
		emb.RestResp = resp
	}

	return emb, nil
}
