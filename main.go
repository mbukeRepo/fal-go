package main

import (
	"context"
	"fal/fal"
	"fmt"
	"time"
)

func GetStatus(status fal.QueueStatus) {

}

func main() {
	client, err := fal.NewClient(fal.WithTokenFromEnv())

	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Second)

	defer func() {
		cancel()
	}()

	res, err := client.Queue.Subscribe(ctx, "fal-ai/fast-sdxl", &fal.QueueSubscribeOptions{
		PollInterval:  500,
		Input:         map[string]interface{}{"prompt": "a cat"},
		OnQueueUpdate: GetStatus,
		Logs:          true,
	})

	if err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", res.(map[string]interface{})["images"].([]interface{})[0].(map[string]interface{})["url"])

}
