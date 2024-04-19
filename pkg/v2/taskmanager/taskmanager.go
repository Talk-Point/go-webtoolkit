package taskmanager

import (
	"context"
	"fmt"
	"os"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	taskspb "cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Task struct {
	Method  string
	Path    string
	Payload []byte
	Queue   string
}

type TaskManager struct {
	Project     string
	Location    string
	BaseURL     string
	AuthKey     string
	TaskHandler func(manager *TaskManager, task Task) error
}

func TestTaskHandler(manager *TaskManager, task Task) error {
	return nil
}

func CloudTaskHandler(manager *TaskManager, task Task) error {
	ctx := context.Background()
	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	endpoint := manager.BaseURL + task.Path

	parent := "projects/" + manager.Project + "/locations/" + manager.Location + "/queues/" + task.Queue
	req := &taskspb.CreateTaskRequest{
		Parent: parent,
		Task: &taskspb.Task{
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
		"X-API-Key":    manager.AuthKey,
	}
	req.Task.GetHttpRequest().Body = []byte(task.Payload)

	testMode := os.Getenv("TEST_MODE")
	if testMode == "true" {
		log.WithFields(log.Fields{
			"task":      endpoint,
			"test-mode": "true",
		}).Debug("Task fired")
	} else {
		_, err = client.CreateTask(ctx, req)
		if err != nil {
			return err
		}
		log.WithFields(log.Fields{
			"task":      endpoint,
			"test-mode": "false",
		}).Debug("Task fired")
	}
	return nil
}

func (m *TaskManager) QueueRemove(queue string) error {
	ctx := context.Background()
	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return err
	}

	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", m.Project, m.Location, queue)
	err = client.DeleteQueue(ctx, &taskspb.DeleteQueueRequest{
		Name: queuePath,
	})
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"queue": queue,
	}).Info("TaskManager Queue removed")
	defer client.Close()
	return nil
}

func (m *TaskManager) QueueCreate(queue string) error {
	ctx := context.Background()
	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", m.Project, m.Location, queue)
	_, err = client.GetQueue(ctx, &taskspb.GetQueueRequest{Name: queuePath})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// The queue does not exist, create it
			_, err = client.CreateQueue(ctx, &taskspb.CreateQueueRequest{
				Parent: fmt.Sprintf("projects/%s/locations/%s", m.Project, m.Location),
				Queue: &taskspb.Queue{
					Name: queuePath,
					RateLimits: &taskspb.RateLimits{
						MaxDispatchesPerSecond:  1,
						MaxConcurrentDispatches: 1,
					},
					RetryConfig: &taskspb.RetryConfig{
						MaxAttempts: 5,
					},
				},
			})
			if err != nil {
				return err
			}
			log.WithFields(log.Fields{
				"queue": queue,
			}).Info("TaskManager Queue created")
		} else {
			return err
		}
	}

	return nil
}

func (m *TaskManager) AddTask(task *Task) error {
	err := m.TaskHandler(m, *task)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"queue":    task.Queue,
			"location": m.Location,
			"project":  m.Project,
		}).Error("TaskManager.AddTask")
	}
	return err
}

func (m *TaskManager) AddTaskAndCreateQueueWhenNotExists(task *Task) error {
	err := m.AddTask(task)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"queue":    task.Queue,
			"location": m.Location,
			"project":  m.Project,
		}).Info("TaskManager.AddTaskAndCreateQueueWhenNotExists but retry")
		err := m.QueueCreate(task.Queue)
		if err != nil {
			return err
		}
		return m.AddTask(task)
	}
	return err
}

type Options struct {
	Project  string
	Location string
	BaseURL  string
	AuthKey  string
}

func NewCloudTaskManager(opts *Options) *TaskManager {
	return &TaskManager{
		Project:     opts.Project,
		Location:    opts.Location,
		BaseURL:     opts.BaseURL,
		AuthKey:     opts.AuthKey,
		TaskHandler: CloudTaskHandler,
	}
}

func NewTestTaskManager() *TaskManager {
	return &TaskManager{
		Project:     "test",
		Location:    "test",
		BaseURL:     "http://localhost:8080",
		AuthKey:     "test",
		TaskHandler: TestTaskHandler,
	}
}
