package main

import (
	"context"
	"fal/fal"
)

func GetStatus(status fal.QueueStatus) {

}

func main() {
	client, err := fal.NewClient(fal.WithTokenFromEnv())

	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		cancel()
	}()

	client.Queue.Subscribe(ctx, "fal-ai/fast-lightning-sdxl", &fal.QueueSubscribeOptions{
		PollInterval:  500,
		Input:         map[string]interface{}{"text": "Hello, World!"},
		OnQueueUpdate: GetStatus,
	})

}
