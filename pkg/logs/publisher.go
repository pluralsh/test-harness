package logs

import (
	"fmt"
	phx "github.com/Douvi/gophoenix"
	testv1alpha1 "github.com/pluralsh/test-harness/api/v1alpha1"
	"github.com/pluralsh/test-harness/pkg/plural"
	"strings"
	"sync"
)

type LogPublisher struct {
	Client  *plural.Client
	Test    *testv1alpha1.TestSuite
	Channel *phx.Channel
	Buffer  map[string][]string
	Wait    *sync.WaitGroup
}

type LogMessage struct {
	Line string `json:"line"`
	Id   string `json:"step"`
}

const flushLen = 10

func NewPublisher(mgr *LogManager, test *testv1alpha1.TestSuite) *LogPublisher {
	return &LogPublisher{
		Client: plural.NewUploadClient(mgr.Config),
		Test:   test,
		Buffer: make(map[string][]string),
		Wait:   &sync.WaitGroup{},
	}
}

func (pub *LogPublisher) Publish(line string, step *testv1alpha1.StepStatus) error {
	fmt.Printf("Publishing %s\n", line)
	id := step.PluralId
	buf, ok := pub.Buffer[id]
	if !ok {
		buf = []string{}
	}
	pub.Buffer[id] = append(buf, line)

	if len(pub.Buffer[id]) >= flushLen {
		return pub.deliver(id)
	}

	return nil
}

func (pub *LogPublisher) Close() error {
	for id := range pub.Buffer {
		if len(pub.Buffer[id]) > 0 {
			if err := pub.deliver(id); err != nil {
				return err
			}
		}
	}

	return nil
}

func (pub *LogPublisher) deliver(id string) error {
	buf, _ := pub.Buffer[id]
	logs := strings.Join(buf, "\n")
	pub.Buffer[id] = []string{}
	return pub.Client.PublishLogs(id, logs)
}
