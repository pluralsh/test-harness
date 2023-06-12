package plural

import (
	"fmt"
	"os"

	"github.com/pluralsh/gqlclient"
	"github.com/pluralsh/gqlclient/pkg/utils"
)

type Status string

const (
	StatusQueued    Status = "QUEUED"
	StatusRunning   Status = "RUNNING"
	StatusSucceeded Status = "SUCCEEDED"
	StatusFailed    Status = "FAILED"
)

type TestStep struct {
	Id          string `json:"id,omitempty"`
	Name        string
	Description string
	Status      Status
}

type Test struct {
	Id         string `json:"id,omitempty"`
	Status     Status
	Name       string
	PromoteTag string
	Tags       []string
	Steps      []*TestStep
}

func (client *Client) CreateTest(repo string, test gqlclient.TestAttributes) (*Test, error) {
	resp, err := client.pluralClient.CreateTest(client.ctx, repo, test)
	if err != nil {
		return nil, err
	}

	return convertTest(resp.CreateTest)
}

func (client *Client) UpdateTest(id string, test gqlclient.TestAttributes) (*Test, error) {
	resp, err := client.pluralClient.UpdateTest(client.ctx, id, test)
	if err != nil {
		return nil, err
	}

	return convertTest(resp.UpdateTest)
}

func (client *Client) PublishLogs(stepId, logs string) error {
	_, err := client.pluralClient.PublishLogs(client.ctx, stepId, logs)
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) UpdateStep(id string, logFile string) error {
	f, err := os.Open(logFile)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = client.pluralClient.UpdateStep(client.ctx, id, "logs", gqlclient.WithFiles([]gqlclient.Upload{
		{
			Field: "logs",
			Name:  logFile,
			R:     f,
		},
	}))
	if err != nil {
		return err
	}
	return nil
}

func convertTest(testFragment *gqlclient.TestFragment) (*Test, error) {
	if testFragment == nil {
		return nil, fmt.Errorf("the Test response is nil")
	}
	t := &Test{
		Id:         testFragment.ID,
		Status:     Status(testFragment.Status),
		Name:       utils.ConvertStringPointer(testFragment.Name),
		PromoteTag: testFragment.PromoteTag,
		Steps:      []*TestStep{},
	}
	for _, step := range testFragment.Steps {
		s := &TestStep{
			Id:          step.ID,
			Name:        step.Name,
			Description: step.Description,
			Status:      Status(step.Status),
		}
		t.Steps = append(t.Steps, s)
	}
	return t, nil
}
