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
	mu      sync.Mutex
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
	pub.mu.Lock()
	defer pub.mu.Unlock()
	id := step.PluralId
	buf, ok := pub.Buffer[id]
	if !ok {
		buf = make([]string, 0, flushLen)
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
	pub.Buffer[id] = make([]string, 0, flushLen)
	fmt.Println("publishing log batch for ", id)
	return pub.Client.PublishLogs(id, logs)
}
