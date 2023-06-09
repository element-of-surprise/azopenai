// Package chat details REST messages used in the Chat API.
package chat

import (
	"github.com/element-of-surprise/azopenai/rest/messages/custom"
)

// Req represents a request to the chat API.
type Req struct {
	// Messages to generate chat completions for, in the chat format.
	Messages []SendMsg `json:"messages"`

	// Stop provides up to 4 sequences where the API will stop generating further tokens.
	Stop []string `json:"stop,omitempty"`

	// LogitBias is the likelihood of specified tokens appearing in the completion.
	// This maps tokens (specified by their token ID in the GPT tokenizer) to an associated bias value from -100 to 100.
	// You can use this tokenizer tool (which works for both GPT-2 and GPT-3) to convert text to token IDs.
	// Mathematically, the bias is added to the logits generated by the model prior to sampling.
	// The exact effect will vary per model, but values between -1 and 1 should decrease or increase likelihood of selection;
	// values like -100 or 100 should result in a ban or exclusive selection of the relevant token.
	// As an example, you can pass {\"50256\" &#58; -100} to prevent the <|endoftext|> token from being generated.
	LogitBias map[string]float64 `json:"logit_bias,omitempty"`

	// User is a unique identifier representing your end-user, which can help monitoring and detecting abuse.
	User string `json:"user,omitempty"`

	// N is the number of completions to generate for each prompt. Minimum of 1 and maximum of 128 allowed.
	// Note: Because this parameter generates many completions, it can quickly consume your token quota.
	// Use carefully and ensure that you have reasonable settings for MaxTokens and stop.
	N int `json:"n,omitempty"`

	// MaxTokens is the token count of your prompt. This cannot exceed the model's context length.
	// Most models have a context length of 2048 tokens (except for the newest models, which support 4096). Has minimum of 0.
	MaxTokens int `json:"max_tokens,omitempty"`

	// Temperature is the sampling temperature to use. Higher values means the model will take more risks.
	// Try 0.9 for more creative applications, and 0 (argmax sampling) for ones with a well-defined answer.
	// It is generally recommend altering this or TopP but not both.
	Temperature float64 `json:"temperature,omitempty"`

	// TopP is an alternative to sampling with temperature, called nucleus sampling.
	// This is where the model considers the results of the tokens with TopP probability mass.
	// So 0.1 means only the tokens comprising the top 10% probability mass are considered.
	// It is generally recommend altering this or temperature but not both.
	TopP float64 `json:"top_p,omitempty"`

	// PresencePenalty is a float64 between -2.0 and 2.0. Positive values penalize new tokens based on
	// whether they appear in the text so far, increasing the model's likelihood to talk about new topics.
	PresencePenalty float64 `json:"presence_penalty,omitempty"`

	// FrequencyPenalty is a float64 between -2.0 and 2.0. Positive values penalize new tokens based on their
	// existing frequency in the text so far, decreasing the model's likelihood to repeat the same line verbatim.
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`

	// Stream indicates whether to stream back partial progress. If set, tokens will be sent as data-only server-sent
	// events as they become available, with the stream terminated by a data: [DONE] message.
	Stream bool `json:"stream,omitempty"`
}

// Defaults sets the default values for the request. You must do this before settings
// any values to avoid overwriting fields you set.
func (c Req) Defaults() Req {
	c.Temperature = 1
	c.TopP = 1
	c.N = 1
	c.MaxTokens = 4096
	return c
}

// Role is a the type of role of the author of a message.
type Role string

const (
	// UnknownRole is the default value for Role, indicating that the role was not set.
	UnknownRole Role = ""
	// User is a user in a chat.
	User Role = "user"
	// System is a system message.
	System Role = "system"
	// Assistant is an assistant message.
	Assistant Role = "assistant"
)

// SendMsg is a message to send to the chat API.
type SendMsg struct {
	// Role of the author of this message.
	Role Role `json:"role"`

	// Contents of the message.
	Content string `json:"content"`

	// Name of the user in chat.
	Name string `json:"name,omitempty"`
}

// Resp is the response from the chat API.
type Resp struct {
	// ID is the ID of the chat request.
	ID string `json:"id"`
	// Object is the type of object, such as "chat.completion"
	Object string `json:"object"`
	// Created is the time the chat request was created.
	Created custom.UnixTime `json:"created"`
	// Model is the model used for the chat request, such as "gpt-35-turbo".
	Model string `json:"model"`
	// Choices is the list of chat completions.
	Choices []Choice `json:"choices"`
	// Usage is usage information for the chat request.
	Usage Usage `json:"usage"`
}

// Choice is a chat completion.
type Choice struct {
	// Index is the index of the prompt that this completion corresponds to.
	Index int `json:"index"`
	// Message is the message received from the chat API.
	Message RecvMsg `json:"message"`
	// FinishReason is the reason the chat session ended.
	FinishReason string `json:"finish_reason"`
}

// RecvMsg is a message received from the chat API.
type RecvMsg struct {
	// Role is the role of the author of this message.
	Role Role `json:"role"`
	// Content is the content of the message.
	Content string `json:"content"`
}

// Usage is the usage information for a chat request.
type Usage struct {
	// PromptTokens is the number of tokens used for the prompt.
	PromptTokens int `json:"prompt_tokens"`
	// CompletionTokens is the number of tokens used for the completion.
	CompletionTokens int `json:"completion_tokens"`
	// Tokens is the total number of tokens used.
	TotalTokens int `json:"total_tokens"`
}
