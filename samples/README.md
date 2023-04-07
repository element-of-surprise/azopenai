# AzOpenAI Go SDK Samples

Samples and tests for the [AzOpenAI Go SDK](https://github.com/element-of-surprise/azopenai).

## Usage

Deploy Azure OpenAI Service with models `gpt-35-turbo` (Chat), `text-davinci-003` (Completions), `text-embedding-ada-002` (Embeddings). These are currently hard-coded in [main_test.go](./main_test.go).

Explore functions in [main.go](./main.go).

Get API key and set environment variables.

```bash
export API_KEY='...'
export RESOURCE_NAME='openai230300'
export DEPLOYMENT_ID='gpt-35-turbo'
```

Run sample.

```bash
go run .
```

Run all tests.

```bash
go test
```

Run single test.

```bash
go test -v run TestChat
```
