// Package auth provides the authorization options for authenticating to the Azure
// OpenAi service.
package auth

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

const (
	unknown = iota
	useApiKey
	useAzIdentity
)

// Authorizer provides authorization options for authenticating to the Azure service.
type Authorizer struct {
	// ApiKey provides authentication/authorization using an API key.
	ApiKey string
	// AzIdentity provides authentication/authorization using the AzIdentity package.
	AzIdentity AzIdentity

	method int
}

// Validate validates the Authorizer has the required fields.
func (a Authorizer) Validate() (Authorizer, error) {
	if reflect.ValueOf(a).IsZero() {
		return Authorizer{}, fmt.Errorf("Authorizer must have ApiKey or AzIdentity set")
	}

	if a.ApiKey != "" {
		a.method = useApiKey
		return a, nil
	}
	if err := a.AzIdentity.validate(); err != nil {
		return Authorizer{}, err
	}
	a.method = useAzIdentity
	return a, nil
}

// Authorize adds the authorization header to the request.
func (a Authorizer) Authorize(ctx context.Context, req *http.Request) error {
	if a.method == unknown {
		return fmt.Errorf("unknown authorization method")
	}

	if a.method == useApiKey {
		req.Header.Add("api-key", a.ApiKey)
		return nil
	}

	t, err := a.AzIdentity.Credential.GetToken(ctx, a.AzIdentity.Policy)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", t))
	req.Header.Add("Content-Type", "application/json")
	return err
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
