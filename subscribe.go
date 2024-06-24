package fal

import (
	"context"
)

type Status int

const (
	IN_PROGRESS Status = iota
	COMPLETED
	IN_QUEUE
)

type Method int

const (
	GET Method = iota
	POST
	PUT
	DELETE
)

type QueueStatus struct {
	Status        `json:"status" binding:"oneof=in_progress completed in_queue"`
	QueuePosition int    `json:"queue_position"`
	ResponseUrl   string `json:"response_url"`
}

type QueueSubscribeOptions struct {
	PollInterval  int
	OnEnqueue     *func(requestId string)
	WebhookUrl    string
	Logs          bool
	OnQueueUpdate *func(status QueueStatus)
}

type RunOptions struct {
	Path       string
	Input      interface{}
	AutoUpload bool
	Method
}

type EnqueueResult struct {
	RequestId string `json:"request_id"`
}

type Queue struct {
	c Client
}

func (q *Queue) Subscribe(options RunOptions) (string, error) {
	return "", nil
}

func (q *Queue) GetStatus(requestId string) (QueueStatus, error) {
	return QueueStatus{}, nil
}

func (q *Queue) GetResponse(requestId string) (interface{}, error) {
	return nil, nil
}

func (q *Queue) Submit(ctx context.Context, requestId string) (*EnqueueResult, error) {
	var out interface{}
	err := q.c.Fetch(ctx, "POST", "/queue/submit", map[string]string{"request_id": requestId}, &out)
	if err != nil {
		return nil, err
	}

	return &EnqueueResult{RequestId: requestId}, nil
}

func Subscribe(id string, options QueueSubscribeOptions) {
	if options.OnEnqueue != nil {
		(*options.OnEnqueue)(id)
	}
}
