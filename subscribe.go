package fal

// TODO: double check the appId, the method and the path with the post request data made
// TODO: comment everything and write some tests
import (
	"context"
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
	Options *UrlOptions
}

type EnqueueResult struct {
	RequestId   string `json:"request_id"`
	ResponseUrl string `json:"response_url"`
	StatusUrl   string `json:"status_url"`
	CancelUrl   string `json:"cancel_url"`
}

type QueueResult struct {
	Status   `json:"status" binding:"oneof=in_progress completed in_queue"`
	Logs     *interface{} `json:"logs"`
	Response *interface{} `json:"response"`
}

type Queue struct {
	c         Client
	Subdomain string
}

func NewQueue(c Client, subdomain string) *Queue {
	return &Queue{c: c, Subdomain: subdomain}
}

func (q *Queue) Subscribe(ctx context.Context, id string, runOptions *QueueSubscribeOptions) (*interface{}, error) {
	if runOptions.OnEnqueue != nil {
		(*runOptions.OnEnqueue)(id)
	}

	result, err := q.Submit(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	resultChannel := make(chan interface{})
	errorChannel := make(chan error)
	stopChannel := make(chan struct{})
	requestIdChan := make(chan string, 1)
	requestIdChan <- result.RequestId

	go func() {
		ticker := time.NewTicker(time.Duration(runOptions.PollInterval) * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stopChannel:
				return
			case <-ticker.C:
				status, err := q.GetStatus(ctx, <-requestIdChan, &RunOptions{
					Path: "/status",
					Options: &UrlOptions{
						AppId: id,
					},
				})
				if err != nil {
					errorChannel <- err
					close(stopChannel)
					return
				}

				if runOptions.OnQueueUpdate != nil {
					(*runOptions.OnQueueUpdate)(*status)
				}

				if status.Status == COMPLETED {
					result, err := q.Result(ctx, <-requestIdChan, &RunOptions{})

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

	return nil, nil
}

func (q *Queue) Result(ctx context.Context, requestId string, runOptions *RunOptions) (*QueueResult, error) {
	var out interface{}
	err := q.c.Fetch(ctx, string(GET), runOptions.Path, nil, &out, runOptions.Options)
	if err != nil {
		return nil, err
	}

	return out.(*QueueResult), nil
}

func (q *Queue) GetStatus(ctx context.Context, requestId string, runOptions *RunOptions) (*QueueStatus, error) {
	var out interface{}
	err := q.c.Fetch(ctx, string(GET), runOptions.Path, nil, &out, runOptions.Options)
	if err != nil {
		return nil, err
	}

	return out.(*QueueStatus), nil
}

func (q *Queue) Submit(ctx context.Context, requestId string, runOptions *RunOptions) (*EnqueueResult, error) {
	var out interface{}
	err := q.c.Fetch(ctx,
		string(runOptions.Method),
		runOptions.Path,
		map[string]string{"request_id": requestId},
		&out,
		runOptions.Options,
	)

	if err != nil {
		return nil, err
	}

	return out.(*EnqueueResult), nil
}
