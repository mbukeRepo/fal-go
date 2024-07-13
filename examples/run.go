package main

import (
	"context"
	"fmt"

	"github.com/mbukeRepo/fal-go"
)

func main() {
	client, err := fal.NewClient(fal.WithTokenFromEnv())
	if err != nil {
		panic(err)
	}

	res, err := client.Run(context.Background(), "fal-ai/fast-sdxl", &fal.RunInput{
		"prompt": "photo of a rhino dressed suit and tie sitting at a table in a bar with a bar stools, award winning photography, Elke vogelsang",
	})

	if err != nil {
		panic(err)
	}

	fmt.Println(res)
}
