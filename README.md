# The fal.ai GoLang client

The fal serverless GoLant Client is a robust and developer-friendly library designed for seamless integration of fal serverless functions.

### Getting Started

#### Installation

Use go get to install the package:

```bash
go get -u github.com/mbukeRepo/fal-go
```

Include the package in your project:

```go
import "github.com/mbukeRepo/fal-go"
```

#### Usage

Create a client

```go
client, err := fal.NewClient(fal.WithTokenFromEnv())

if err != nil {
	panic(err)
}
ctx, cancel := context.WithCancel(context.Background())

defer func() {
	cancel()
}()
```

Run a model

```go
res, err := client.Queue.Subscribe(ctx, "fal-ai/fast-sdxl", &fal.QueueSubscribeOptions{
	PollInterval:  500,
	Input:         map[string]interface{}{"prompt": "a cat"},
	Logs:          true,
})

if err != nil {
	panic(err)
}
```
