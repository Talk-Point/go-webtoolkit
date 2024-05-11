package taskmanager

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	taskspb "cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	"github.com/golang/protobuf/ptypes/timestamp"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type QueueOptions struct {
	MaxDispatchesPerSecond  float64
	MaxConcurrentDispatches int32
	MaxAttempts             int32
}

type TaskManagerOptions struct {
	Project   string
	Location  string
	AuthKey   string
	BaseUrl   string
	QueueOpts *QueueOptions
}

type TaskManager struct {
	Client  *cloudtasks.Client
	options *TaskManagerOptions
}

func NewTaskManager(options *TaskManagerOptions) (*TaskManager, error) {
	ctx := context.Background()

	tm := &TaskManager{
		options: options,
	}

	if tm.options.QueueOpts == nil {
		tm.options.QueueOpts = &QueueOptions{
			MaxDispatchesPerSecond:  1,
			MaxConcurrentDispatches: 1,
			MaxAttempts:             5,
		}
	}
	if tm.options.QueueOpts.MaxAttempts == 0 {
		tm.options.QueueOpts.MaxAttempts = 5
	}
	if tm.options.QueueOpts.MaxConcurrentDispatches == 0 {
		tm.options.QueueOpts.MaxConcurrentDispatches = 1
	}
	if tm.options.QueueOpts.MaxDispatchesPerSecond == 0 {
		tm.options.QueueOpts.MaxDispatchesPerSecond = 1
	}

	var err error
	tm.Client, err = cloudtasks.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return tm, nil
}

func (tm *TaskManager) Close() error {
	return tm.Client.Close()
}

type TaskOption func(*Task)

func WithPayload(payload []byte) TaskOption {
	return func(t *Task) {
		t.Payload = payload
	}
}

func WithMethod(method string) TaskOption {
	return func(t *Task) {
		t.Method = method
	}
}

func WithDelay(delay time.Duration) TaskOption {
	return func(t *Task) {
		t.Delay = delay
	}
}

type Task struct {
	Queue   string
	Path    string
	Payload []byte
	Method  string
	Delay   time.Duration
}

func (tm *TaskManager) fireTask(task *Task, req *taskspb.CreateTaskRequest) error {
	testMode := os.Getenv("TEST_MODE")
	if testMode == "true" {
		log.WithFields(log.Fields{
			"queue":   task.Queue,
			"path":    task.Path,
			"method":  task.Method,
			"payload": string(task.Payload),
			"delay":   task.Delay,
		}).Info("TaskManager:run [TESTMODE]")
	} else {
		rv, err := tm.Client.CreateTask(context.Background(), req)
		if err != nil {
			return err
		}
		fmt.Println(rv)
		log.WithFields(log.Fields{
			"queue":   task.Queue,
			"path":    task.Path,
			"method":  task.Method,
			"payload": string(task.Payload),
			"delay":   task.Delay,
		}).Info("TaskManager:run")
	}
	return nil
}

func (tm *TaskManager) Run(queue, path string, options ...TaskOption) error {
	task := &Task{
		Queue:  queue,
		Path:   path,
		Method: "POST",
		Delay:  0,
	}

	for _, option := range options {
		option(task)
	}

	if task.Delay < 0 {
		return errors.New("delay must be greater than or equal to 0")
	}

	baseUrl, err := url.Parse(tm.options.BaseUrl)
	if err != nil {
		log.Fatal(err)
	}

	taskPath, err := url.Parse(task.Path)
	if err != nil {
		log.Fatal(err)
	}

	endpoint := baseUrl.ResolveReference(taskPath).String()

	var t *timestamp.Timestamp
	if task.Delay > 0 {
		t = &timestamp.Timestamp{
			Seconds: time.Now().UTC().Add(task.Delay).Unix(),
			Nanos:   0,
		}
	}

	parent := "projects/" + tm.options.Project + "/locations/" + tm.options.Location + "/queues/" + task.Queue
	req := &taskspb.CreateTaskRequest{
		Parent: parent,
		Task: &taskspb.Task{
			ScheduleTime: t,
			MessageType: &taskspb.Task_HttpRequest{
				HttpRequest: &taskspb.HttpRequest{
					HttpMethod: taskspb.HttpMethod_POST,
					Url:        endpoint,
				},
			},
		},
	}
	req.Task.GetHttpRequest().Headers = map[string]string{
		"Content-Type": "application/json",
		"X-API-Key":    tm.options.AuthKey,
	}
	req.Task.GetHttpRequest().Body = []byte(task.Payload)

	err = tm.fireTask(task, req)
	if err != nil {
		s, ok := status.FromError(err)
		if ok && s.Code() == codes.FailedPrecondition && s.Message() == "Queue does not exist." {
			err = tm.CreateQueue(task.Queue)
			if err != nil {
				return err
			}
			return tm.fireTask(task, req)
		} else {
			return err
		}
	}
	return nil
}

func (tm *TaskManager) CreateQueue(name string) error {
	parent := "projects/" + tm.options.Project + "/locations/" + tm.options.Location
	queue := &taskspb.Queue{
		Name: parent + "/queues/" + name,
		RateLimits: &taskspb.RateLimits{
			MaxDispatchesPerSecond:  tm.options.QueueOpts.MaxDispatchesPerSecond,
			MaxConcurrentDispatches: tm.options.QueueOpts.MaxConcurrentDispatches,
		},
		RetryConfig: &taskspb.RetryConfig{
			MaxAttempts: tm.options.QueueOpts.MaxAttempts,
		},
	}
	_, err := tm.Client.CreateQueue(context.Background(), &taskspb.CreateQueueRequest{
		Parent: parent,
		Queue:  queue,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"queue": name,
		}).Error("TaskManager:CreateQueue errored")
		return err
	}
	log.WithFields(log.Fields{
		"queue": name,
	}).Info("TaskManager:CreateQueue success")
	return nil
}

func (tm *TaskManager) PauseQueue(name string) error {
	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", tm.options.Project, tm.options.Location, name)
	_, err := tm.Client.PauseQueue(context.Background(), &taskspb.PauseQueueRequest{
		Name: queuePath,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"queue": name,
		}).Error("TaskManager:PauseQueue errored")
		return err
	}
	log.WithFields(log.Fields{
		"queue": name,
	}).Debug("TaskManager:PauseQueue success")
	return nil
}

func (tm *TaskManager) ResumeQueue(name string) error {
	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", tm.options.Project, tm.options.Location, name)
	_, err := tm.Client.ResumeQueue(context.Background(), &taskspb.ResumeQueueRequest{
		Name: queuePath,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"queue": name,
		}).Error("TaskManager:ResumeQueue errored")
		return err
	}
	log.WithFields(log.Fields{
		"queue": name,
	}).Debug("TaskManager:ResumeQueue success")
	return nil
}

func (tm *TaskManager) DeleteQueue(name string) error {
	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", tm.options.Project, tm.options.Location, name)
	err := tm.Client.DeleteQueue(context.Background(), &taskspb.DeleteQueueRequest{
		Name: queuePath,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"queue": name,
		}).Error("TaskManager:DeleteQueue errored")
		return err
	}
	log.WithFields(log.Fields{
		"queue": name,
	}).Debug("TaskManager:DeleteQueue success")
	return nil
}

func (tm *TaskManager) GetQueue(name string) (*taskspb.Queue, error) {
	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", tm.options.Project, tm.options.Location, name)
	queue, err := tm.Client.GetQueue(context.Background(), &taskspb.GetQueueRequest{
		Name: queuePath,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"queue": name,
		}).Error("TaskManager:GetQueue errored")
		return nil, err
	}
	log.WithFields(log.Fields{
		"error": err,
		"queue": name,
	}).Debug("TaskManager:GetQueue success")
	return queue, nil
}

func (tm *TaskManager) ListQueues() ([]*taskspb.Queue, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", tm.options.Project, tm.options.Location)
	iter := tm.Client.ListQueues(context.Background(), &taskspb.ListQueuesRequest{
		Parent: parent,
	})
	var queues []*taskspb.Queue
	for {
		queue, err := iter.Next()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.WithFields(log.Fields{
				"error": err,
			}).Error("TaskManager:ListQueues errored")
			return nil, err
		}
		queues = append(queues, queue)
	}
	log.Debug("TaskManager:ListQueues success")
	return queues, nil
}

func (tm *TaskManager) PurgeQueue(name string) error {
	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", tm.options.Project, tm.options.Location, name)
	_, err := tm.Client.PurgeQueue(context.Background(), &taskspb.PurgeQueueRequest{
		Name: queuePath,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"queue": name,
		}).Error("TaskManager:PurgeQueue errored")
		return err
	}
	log.WithFields(log.Fields{
		"queue": name,
	}).Debug("TaskManager:PurgeQueue success")
	return nil
}

func (tm *TaskManager) UpdateQueue(name string, opts *QueueOptions) error {
	queue, err := tm.GetQueue(name)
	if err != nil {
		return err
	}
	if opts.MaxDispatchesPerSecond >= 0 {
		queue.RateLimits.MaxDispatchesPerSecond = opts.MaxDispatchesPerSecond
	}
	if opts.MaxConcurrentDispatches >= 0 {
		queue.RateLimits.MaxConcurrentDispatches = opts.MaxConcurrentDispatches
	}
	if opts.MaxAttempts >= 0 {
		queue.RetryConfig.MaxAttempts = opts.MaxAttempts
	}

	_, err = tm.Client.UpdateQueue(context.Background(), &taskspb.UpdateQueueRequest{
		Queue: queue,
	})

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"queue": name,
		}).Error("TaskManager:UpdateQueue errored")
		return err
	}
	log.WithFields(log.Fields{
		"queue": name,
	}).Debug("TaskManager:UpdateQueue success")
	return nil
}
