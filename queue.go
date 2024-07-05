package fal

import (
	"context"
	"strings"
	"time"
)

type Status string

const (
	IN_PROGRESS Status = "in_progress"
	COMPLETED   Status = "completed"
	IN_QUEUE    Status = "in_queue"
)

type Method string

const (
	GET    Method = "GET"
	POST   Method = "POST"
	PUT    Method = "PUT"
	DELETE Method = "DELETE"
)

type QueueStatus struct {
	Status        `json:"status" binding:"oneof=in_progress completed in_queue"`
	QueuePosition int    `json:"queue_position"`
	ResponseUrl   string `json:"response_url"`
}

type QueueSubscribeOptions struct {
	PollInterval  int
	OnEnqueue     func(requestId string)
	WebhookUrl    string
	Logs          bool
	OnQueueUpdate func(status QueueStatus)
	Input         interface{}
}

type RunOptions struct {
	Path       string
	Input      interface{}
	AutoUpload bool
	Method
	Options *UrlOptions
}

type EnqueueResult struct {
	RequestId   string `json:"request_id"`
	ResponseUrl string `json:"response_url"`
	StatusUrl   string `json:"status_url"`
	CancelUrl   string `json:"cancel_url"`
}

type Queue struct {
	c         *Client
	Subdomain string
}

func (q *Queue) Subscribe(ctx context.Context, id string, runOptions *QueueSubscribeOptions) (interface{}, error) {
	if runOptions.OnEnqueue != nil {
		(runOptions.OnEnqueue)(id)
	}

	result, err := q.Submit(ctx, id, &RunOptions{
		Input: runOptions.Input,
		Path:  "/",
		Options: &UrlOptions{
			Subdomain: "queue",
			AppId:     id,
		},
		Method: POST,
	})
	if err != nil {
		return nil, err
	}
	resultChannel := make(chan interface{})
	errorChannel := make(chan error)
	stopChannel := make(chan struct{})
	requestIdChan := make(chan string, 1)
	appIdChan := make(chan string, 1)
	requestIdChan <- result.RequestId
	appIdChan <- id

	go func() {
		requestId := <-requestIdChan
		appId := <-appIdChan
		for {
			select {
			case <-stopChannel:
				close(stopChannel)
				return
			case <-ctx.Done():
				close(stopChannel)
				return
			default:
				time.Sleep(time.Duration(runOptions.PollInterval) * time.Millisecond)
				status, err := q.GetStatus(ctx, requestId, &RunOptions{
					Path:   "/requests/" + requestId + "/status",
					Method: GET,
					Options: &UrlOptions{
						AppId:     appId,
						Subdomain: "queue",
					},
				})
				if err != nil {
					errorChannel <- err
					close(stopChannel)
					return
				}

				if runOptions.OnQueueUpdate != nil {
					(runOptions.OnQueueUpdate)(*status)
				}

				status.Status = Status(strings.TrimSpace(string(status.Status)))
				if status.Status == Status(strings.ToUpper(string(COMPLETED))) {
					result, err := q.Result(ctx, requestId, &RunOptions{
						Path:   "/requests/" + requestId,
						Method: GET,
						Options: &UrlOptions{
							AppId:     appId,
							Subdomain: "queue",
						},
					})
					if err != nil {
						errorChannel <- err
					} else {
						resultChannel <- result
					}
					close(stopChannel)
					return
				}
			}
		}
	}()

	select {
	case result := <-resultChannel:
		return result, nil
	case err := <-errorChannel:
		return nil, err
	case <-ctx.Done():
		close(stopChannel)
		return nil, ctx.Err()
	}
}

func (q *Queue) Result(ctx context.Context, requestId string, runOptions *RunOptions) (interface{}, error) {
	var out interface{}
	err := q.c.Fetch(ctx, string(GET), runOptions.Path, nil, &out, runOptions.Options)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (q *Queue) GetStatus(ctx context.Context, requestId string, runOptions *RunOptions) (*QueueStatus, error) {
	var out QueueStatus
	err := q.c.Fetch(ctx, string(runOptions.Method), runOptions.Path, nil, &out, runOptions.Options)

	if err != nil {
		return nil, err
	}

	return &out, nil
}

func (q *Queue) Submit(ctx context.Context, requestId string, runOptions *RunOptions) (*EnqueueResult, error) {
	var out EnqueueResult
	err := q.c.Fetch(ctx,
		string(runOptions.Method),
		runOptions.Path,
		runOptions.Input,
		&out,
		runOptions.Options,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
