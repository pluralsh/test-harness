package plural

import (
	"fmt"
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
	Logs        string
}

type Test struct {
	Id         string `json:"id,omitempty"`
	Status     Status
	PromoteTag string
	Steps      []*TestStep
}

const TestFragment = `
	fragment TestFragment on Test {
		id
		status
		promoteTag
		steps {
			id
			name
			description
			status
			logs
		}
	}
`

var createTest = fmt.Sprintf(`
	mutation Create($name: String!, $attrs: TestAttributes!) {
		createTest(name: $name, attributes: $attrs) {
			...TestFragment
		}
	}
	%s
`, TestFragment)

var updateTest = fmt.Sprintf(`
	mutation Update($id: ID!, $attrs: TestAttributes!) {
		updateTest(id: $id, attributes: $attrs) {
			...TestFragment
		}
	}
	%s
`, TestFragment)

func (client *Client) CreateTest(repo string, test *Test) (result *Test, err error) {
	var resp struct {
		CreateTest *Test
	}
	req := client.Build(createTest)
	req.Var("name", repo)
	req.Var("attrs", test)
	err = client.Run(req, &resp)
	result = resp.CreateTest
	return
}

func (client *Client) UpdateTest(test *Test) (result *Test, err error) {
	var resp struct {
		UpdateTest *Test
	}
	req := client.Build(updateTest)
	req.Var("id", test.Id)
	test.Id = "" // hack to bypass serialization
	req.Var("attrs", test)
	err = client.Run(req, &resp)
	result = resp.UpdateTest
	return
}
