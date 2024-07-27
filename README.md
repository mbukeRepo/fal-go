# The fal.ai GoLang client

The fal serverless GoLang Client is a robust and developer-friendly library designed for seamless integration of fal serverless.

### Getting Started

#### Installation

Use go get to install the package:

```bash
go get -u github.com/mbukeRepo/fal-go
```

#### Usage

Start by configuring your credentials:

```
export FAL_AUTH_TOKEN=
```

Include the package in your project:

```go
import "github.com/mbukeRepo/fal-go"
```

Create a client

```go
client, err := fal.NewClient(fal.WithTokenFromEnv())

if err != nil {
	panic(err)
}
```

Running a function with `client.Run`:

```go
ctx, cancel := context.WithCancel(context.Background())

defer func() {
	cancel()
}()

res, err := client.Run(ctx, "fal-ai/fast-sdxl", &fal.RunInput{
		"prompt": "...",
})
```

Long-running functions with `client.Subscribe`:

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
