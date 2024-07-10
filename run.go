package fal

import "context"

type RunInput map[string]interface{}

func (c *Client) Run(ctx context.Context, functionId string, input *RunInput) (*map[string]interface{}, error) {
	var out map[string]interface{}
	err := c.Fetch(ctx, string(POST), functionId, input, &out, nil)
	if err != nil {
		return nil, err
	}

	return &out, nil
}
