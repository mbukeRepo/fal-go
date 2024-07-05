package fal

import "context"

type InferenceOptions struct{}
type InferenceResults struct{}

// TODO: implement single request runs on fal.
func (c *Client) Run(ctx context.Context, client *Client, options *InferenceOptions) (*InferenceResults, error) {
	return nil, nil
}
