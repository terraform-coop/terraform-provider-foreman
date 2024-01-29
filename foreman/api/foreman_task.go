package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"net/http"
	"time"
)

// ForemanTask is either the task from /foreman_tasks/.../<uuid> or a response
// from a Katello endpoint, which uses the async_task (in Foreman source code) function.
// The most important fields are covered, but there are more.
type ForemanTask struct {
	Id        string      `json:"id"`
	Label     string      `json:"label"`
	Pending   bool        `json:"pending"`
	Action    string      `json:"action"`
	Username  string      `json:"username"`
	StartedAt string      `json:"started_at"`
	EndedAt   string      `json:"ended_at"`
	Duration  string      `json:"duration"`
	State     string      `json:"state"`
	Result    string      `json:"result"`
	Progress  float64     `json:"progress"`
	Input     interface{} `json:"input"`
	Output    interface{} `json:"output"`
	Humanized struct {
		Action string      `json:"action"`
		Input  interface{} `json:"input"`
		Output interface{} `json:"output"`
		Errors []string    `json:"errors"`
	} `json:"humanized"`
	CliExample       interface{} `json:"cli_example"`
	StartAt          string      `json:"start_at"`
	AvailableActions struct {
		Cancellable bool `json:"cancellable"`
		Resumable   bool `json:"resumable"`
	} `json:"available_actions"`
}

// waitForKatelloAsyncTask provides a method to wait for a Katello asynchronous task to finish.
func (c *Client) waitForKatelloAsyncTask(taskID string) error {
	log.Tracef("waitForKatelloAsyncTask")

	ctx := context.TODO()
	const endpoint = "/foreman_tasks/api/tasks/%s"
	req, err := c.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(endpoint, taskID), nil)
	if err != nil {
		return err
	}

	// This works as a retry counter, currently set to 3 tries, e.g. 2 retries
	for counter := 0; counter < 3; counter++ {
		log.Tracef("waitForKatelloAsyncTask retry loop with counter %d", counter)

		var task ForemanTask
		err = c.SendAndParse(req, &task)
		if err != nil {
			return err
		}

		log.Debugf("task: %+v", task)
		if !task.Pending {
			return nil
		}

		log.Infof("Task %s is still pending, sleeping for 500ms and then retrying…", task.Id)
		time.Sleep(time.Duration(time.Millisecond * 500))
	}

	// The retries should produce a success. If not, fail with error
	return errors.New("Error in retrying to wait for task " + taskID)
}
