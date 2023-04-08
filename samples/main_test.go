package main

import (
	"os"
	"testing"
)

func TestChat(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	resourceName := os.Getenv("RESOURCE_NAME")
	deploymentID := "gpt-35-turbo"
	if err := Chat(apiKey, resourceName, deploymentID); err != nil {
		t.Fatal(err)
	}
}

func TestChatWithParams(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	resourceName := os.Getenv("RESOURCE_NAME")
	deploymentID := "gpt-35-turbo"
	if err := ChatWithParams(apiKey, resourceName, deploymentID); err != nil {
		t.Fatal(err)
	}
}

func TestChatWithParamsPerCall(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	resourceName := os.Getenv("RESOURCE_NAME")
	deploymentID := "gpt-35-turbo"
	if err := ChatWithParamsPerCall(apiKey, resourceName, deploymentID); err != nil {
		t.Fatal(err)
	}
}

func TestCompletions(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	resourceName := os.Getenv("RESOURCE_NAME")
	deploymentID := "text-davinci-003"
	if err := Completions(apiKey, resourceName, deploymentID); err != nil {
		t.Fatal(err)
	}
}

func TestCompletionsWithParams(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	resourceName := os.Getenv("RESOURCE_NAME")
	deploymentID := "text-davinci-003"
	if err := CompletionsWithParams(apiKey, resourceName, deploymentID); err != nil {
		t.Fatal(err)
	}
}

func TestCompletionsWithParamsPerCall(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	resourceName := os.Getenv("RESOURCE_NAME")
	deploymentID := "text-davinci-003"
	if err := CompletionsWithParamsPerCall(apiKey, resourceName, deploymentID); err != nil {
		t.Fatal(err)
	}
}

func TestEmbeddings(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	resourceName := os.Getenv("RESOURCE_NAME")
	deploymentID := "text-embedding-ada-002"
	if err := Embeddings(apiKey, resourceName, deploymentID); err != nil {
		t.Fatal(err)
	}
}

func TestEmbeddingsWithParams(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	resourceName := os.Getenv("RESOURCE_NAME")
	deploymentID := "text-embedding-ada-002"
	if err := EmbeddingsWithParams(apiKey, resourceName, deploymentID); err != nil {
		t.Fatal(err)
	}
}

func TestEmbeddingsWithParamsPerCall(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	resourceName := os.Getenv("RESOURCE_NAME")
	deploymentID := "text-embedding-ada-002"
	if err := EmbeddingsWithParamsPerCall(apiKey, resourceName, deploymentID); err != nil {
		t.Fatal(err)
	}
}
