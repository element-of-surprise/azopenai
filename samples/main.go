package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/element-of-surprise/azopenai"
	"github.com/element-of-surprise/azopenai/auth"
	"github.com/element-of-surprise/azopenai/clients/chat"
	"github.com/element-of-surprise/azopenai/clients/completions"
	"github.com/element-of-surprise/azopenai/clients/embeddings"
)

func main() {
	apiKey := os.Getenv("API_KEY")
	resourceName := os.Getenv("RESOURCE_NAME")
	deploymentID := os.Getenv("DEPLOYMENT_ID")

	if err := Chat(apiKey, resourceName, deploymentID); err != nil {
		log.Fatal(err)
	}
}

func Chat(apiKey, resourceName, deploymentID string) error {
	client, err := azopenai.New(resourceName, deploymentID, auth.Authorizer{ApiKey: apiKey})
	if err != nil {
		return err
	}

	chatClient := client.Chat()
	messages := []chat.SendMsg{
		{
			Role:    chat.System,
			Content: "You are a helpful assistant.",
		},
		{
			Role:    chat.User,
			Content: "Does Azure OpenAI support customer managed keys?",
		},
	}
	resp, err := chatClient.Call(context.Background(), messages)
	if err != nil {
		return err
	}
	fmt.Println(resp.Text[0])

	return nil
}

func ChatWithParams(apiKey, resourceName, deploymentID string) error {
	client, err := azopenai.New(resourceName, deploymentID, auth.Authorizer{ApiKey: apiKey})
	if err != nil {
		return err
	}

	chatClient := client.Chat()
	// This creates a new instance of CallParams with the default values.
	// We then modify then and set them on the client. They will be used on
	// every call unless you override them on a specific call.
	params := chat.CallParams{}.Defaults()
	params.MaxTokens = 32
	params.Temperature = 0.5
	chatClient.SetParams(params)

	messages := []chat.SendMsg{{Role: chat.User, Content: "Tell me a joke"}}
	resp, err := chatClient.Call(context.Background(), messages)
	if err != nil {
		return err
	}
	fmt.Println(resp.Text[0])

	return nil
}

func ChatWithParamsPerCall(apiKey, resourceName, deploymentID string) error {
	client, err := azopenai.New(resourceName, deploymentID, auth.Authorizer{ApiKey: apiKey})
	if err != nil {
		return err
	}

	chatClient := client.Chat()
	// This creates a new instance of CallParams with the default values.
	// We then modify then and set them on the client. They will be used on
	// every call unless you override them on a specific call.
	params := chat.CallParams{}.Defaults()
	params.MaxTokens = 32
	params.Temperature = 0.5

	messages := []chat.SendMsg{{Role: "user", Content: "Tell me a joke"}}
	resp, err := chatClient.Call(context.Background(), messages, chat.WithCallParams(params))
	if err != nil {
		return err
	}
	fmt.Println(resp.Text[0])

	return nil
}

func Completions(apiKey, resourceName, deploymentID string) error {
	client, err := azopenai.New(resourceName, deploymentID, auth.Authorizer{ApiKey: apiKey})
	if err != nil {
		return err
	}

	completions := client.Completions()
	resp, err := completions.Call(context.Background(), []string{"The capital of California is"})
	if err != nil {
		return err
	}
	fmt.Println(resp.Text[0])
	return nil

}

func CompletionsWithParams(apiKey, resourceName, deploymentID string) error {
	client, err := azopenai.New(resourceName, deploymentID, auth.Authorizer{ApiKey: apiKey})
	if err != nil {
		return err
	}

	completionsClient := client.Completions()

	// This creates a new instance of CallParams with the default values.
	// We then modify then and set them on the client. They will be used on
	// every call unless you override them on a specific call.
	params := completions.CallParams{}.Defaults()
	params.MaxTokens = 32
	params.Temperature = 0.5
	completionsClient.SetParams(params)

	resp, err := completionsClient.Call(context.Background(), []string{"The capital of California is"})
	if err != nil {
		return err
	}
	fmt.Println(resp.Text[0])
	return nil

}

func CompletionsWithParamsPerCall(apiKey, resourceName, deploymentID string) error {
	client, err := azopenai.New(resourceName, deploymentID, auth.Authorizer{ApiKey: apiKey})
	if err != nil {
		return err
	}

	completionsClient := client.Completions()

	// This creates a new instance of CallParams with the default values.
	// We then modify then and set them on the client. They will be used on
	// every call unless you override them on a specific call.
	params := completions.CallParams{}.Defaults()
	params.MaxTokens = 32
	params.Temperature = 0.5

	resp, err := completionsClient.Call(context.Background(), []string{"The capital of California is"}, completions.WithCallParams(params))
	if err != nil {
		return err
	}
	fmt.Println(resp.Text[0])
	return nil

}

func Embeddings(apiKey, resourceName, deploymentID string) error {
	client, err := azopenai.New(resourceName, deploymentID, auth.Authorizer{ApiKey: apiKey})
	if err != nil {
		return err
	}

	embeddingsClient := client.Embeddings()
	text := []string{"The food was delicious and the waiter..."}
	resp, err := embeddingsClient.Call(context.Background(), text)
	if err != nil {
		return err
	}
	fmt.Printf("%v", resp.Results)
	return nil
}

func EmbeddingsWithParams(apiKey, resourceName, deploymentID string) error {
	client, err := azopenai.New(resourceName, deploymentID, auth.Authorizer{ApiKey: apiKey})
	if err != nil {
		return err
	}

	embeddingsClient := client.Embeddings()

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
	return nil
}

func EmbeddingsWithParamsPerCall(apiKey, resourceName, deploymentID string) error {
	client, err := azopenai.New(resourceName, deploymentID, auth.Authorizer{ApiKey: apiKey})
	if err != nil {
		return err
	}

	embeddingsClient := client.Embeddings()

	// This creates a new instance of CallParams with the default values.
	// We then modify then and set them on the client. They will be used on
	// every call unless you override them on a specific call.
	params := embeddings.CallParams{}
	params.User = "element-of-surprise"

	text := []string{"The food was delicious and the waiter..."}
	resp, err := embeddingsClient.Call(context.Background(), text, embeddings.WithCallParams(params))
	if err != nil {
		return err
	}
	fmt.Printf("%v", resp.Results)
	return nil
}
