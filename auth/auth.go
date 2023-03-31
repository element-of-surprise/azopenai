// Package auth provides the authorization options for authenticating to the Azure
// OpenAi service.
package auth

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

// Authorizer provides authorization options for authenticating to the Azure service.
type Authorizer struct {
	// AzIdentity provides authentication/authorization using the AzIdentity package.
	AzIdentity AzIdentity
}

// Validate validates the Authorizer has the required fields.
func (a Authorizer) Validate() error {
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
