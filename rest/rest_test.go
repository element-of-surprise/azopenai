package rest

import (
	"testing"
)

func TestEndpoints(t *testing.T) {
	tests := []struct {
		desc         string
		deploymentID string
		endpointType endpointType
		want         string
	}{
		{
			desc:         "completions",
			deploymentID: "deployment1",
			endpointType: completionsTmpl,
			want:         "https://test.openai.azure.com/openai/deployments/deployment1/completions?api-version=" + APIVersion,
		},
		{
			desc:         "embeddings",
			deploymentID: "deployment1",
			endpointType: embeddingsTmpl,
			want:         "https://test.openai.azure.com/openai/deployments/deployment1/embeddings?api-version=" + APIVersion,
		},
		{
			desc:         "chat",
			deploymentID: "deployment1",
			endpointType: chatTmpl,
			want:         "https://test.openai.azure.com/openai/deployments/deployment1/chat/completions?api-version=" + APIVersion,
		},
		{
			desc:         "completions, but different deployment",
			deploymentID: "deployment2",
			endpointType: completionsTmpl,
			want:         "https://test.openai.azure.com/openai/deployments/deployment2/completions?api-version=" + APIVersion,
		},
		{
			desc:         "completions, checking original deployment still exists",
			deploymentID: "deployment1",
			endpointType: completionsTmpl,
			want:         "https://test.openai.azure.com/openai/deployments/deployment1/completions?api-version=" + APIVersion,
		},
	}
	e := newEndpoints()
	vars := templVars{
		ResourceName: "test",
		APIVersion:   APIVersion,
	}
	for _, test := range tests {
		u, err := e.url(test.endpointType, test.deploymentID, vars)
		if err != nil {
			panic(err)
		}

		if u.String() != test.want {
			t.Errorf("TestURLs(%s): got %s, want %s", test.desc, u.String(), test.want)
		}
	}
}
