<p style="text-align:center">
<a href="https://github.com/element-of-surprise/azopenai"><img src="https://raw.githubusercontent.com/element-of-surprise/azopenai/main/logo.svg?sanitize=true" alt="OpenAI Logo" width="100"></a>
</p> 
<h1 style="text-align:center;">AzOpenAI Go SDK</h1>

[![GoDev](https://img.shields.io/static/v1?label=godev&message=reference&color=00add8)](https://pkg.go.dev/github.com/element-of-surprise/azopenai)
[![Go Report Card](https://goreportcard.com/badge/github.com/element-of-surprise/azopenai)](https://goreportcard.com/report/github.com/element-of-surprise/azopenai)

AzOpenAI is a Go SDK for interfacing with [Azure OpenAI Service](https://learn.microsoft.com/azure/cognitive-services/openai/overview). It provides a simple and easy-to-use interface for requesting various Azure OpenAI Service operations, including [Chat Completions](https://learn.microsoft.com/en-us/azure/cognitive-services/openai/reference#chat-completions), [Completions](https://learn.microsoft.com/en-us/azure/cognitive-services/openai/reference#completions) and [Embeddings](https://learn.microsoft.com/en-us/azure/cognitive-services/openai/reference#embeddings).

# Overview 

[Installation](#installation) |
[Usage](#usage) |
[License](#license) |
[Documentation](https://pkg.go.dev/github.com/element-of-surprise/azopenai) | 
[Azure OpenAI Service](https://azure.microsoft.com/products/cognitive-services/openai-service)

## Installation

To use AzOpenAI in your Go project, add the following import statement.

```go
import "github.com/element-of-surprise/azopenai"
```

Then, run `go get` (after initializing your module with [go mod init](https://go.dev/doc/tutorial/create-module#start)), to download and install the package.

```bash
go get github.com/element-of-surprise/azopenai
```

## Usage

Here is an example of how to use AzOpenAI to generate text using the OpenAI completions API endpoint. 

See [samples/main.go](samples/main.go) for full Completions, Chat and Embeddings samples.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/element-of-surprise/azopenai"
	"github.com/element-of-surprise/azopenai/auth"
)

func main() {
	ctx := context.Background()
	apiKey := os.Getenv("API_KEY")
	resourceName := os.Getenv("RESOURCE_ID")
	deploymentID := os.Getenv("DEPLOYMENT_ID")

	client, err := azopenai.New(resourceName, auth.Authorizer{ApiKey: apiKey})
	if err != nil {
		log.Fatal(err)
	}

	completions := client.Completions(deploymentID)

	resp, err := completions.Call(ctx, []string{"The capital of California is"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Text[0])
}
```

In this example, `azopenai.New()` is used to create a new AzOpenAI client with your Azure OpenAI Service [API Key](https://learn.microsoft.com/en-us/azure/cognitive-services/openai/reference#authentication). 

Then the returned client aggregator `client` has `client.Completions()` called to get a client for the completions endpoint. You pass the `deploymentID` here so that you can point at the right model for the sub-client. Deployments only have support for some API calls.

Next, `completions.Call()` is invoked on the client with the prompt(s) and any additional options specified. 

Finally, response returned by the Completions endpoint is printed to the console.

## License

AzOpenAI Go SDK is licensed under the MIT License. See LICENSE for more information.

## More information
For more detailed information on the SDK, please see the [![GoDev](https://img.shields.io/static/v1?label=godev&message=reference&color=00add8)](https://pkg.go.dev/github.com/element-of-surprise/azopenai) which includes more authentication examples, using other API endpoints, details on options, and more.