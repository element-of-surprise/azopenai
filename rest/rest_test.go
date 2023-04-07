package rest

import (
	"testing"
)

func TestURLs(t *testing.T) {
	c := Client{
		vars: templVars{
			ResourceName: "test",
			DeploymentID: "test",
			APIVersion:   APIVersion,
		},
	}

	if err := c.urls(); err != nil {
		t.Fatal(err)
	}

	if c.completionsURL.String() != "https://test.openai.azure.com/openai/deployments/test/completions?api-version="+APIVersion {
		t.Fatalf("unexpected completions url: %s", c.completionsURL.String())
	}

	if c.embeddingsURL.String() != "https://test.openai.azure.com/openai/deployments/test/embeddings?api-version="+APIVersion {
		t.Fatalf("unexpected embeddings url: %s", c.embeddingsURL.String())
	}

	if c.chatURL.String() != "https://test.openai.azure.com/openai/deployments/test/chat?api-version="+APIVersion {
		t.Fatalf("unexpected chat url: %s", c.chatURL.String())
	}
}
