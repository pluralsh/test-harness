package logs

import (
	"fmt"
	phx "github.com/Douvi/gophoenix"
	testv1alpha1 "github.com/pluralsh/test-harness/api/v1alpha1"
	"github.com/pluralsh/test-harness/pkg/plural"
)

type LogPublisher struct {
	Socket  *plural.Socket
	Test    *testv1alpha1.TestSuite
	Channel *phx.Channel
	Open    bool
}

type LogMessage struct {
	Line string
	Id   string
}

func NewPublisher(socket *plural.Socket, test *testv1alpha1.TestSuite) *LogPublisher {
	return &LogPublisher{Socket: socket, Test: test}
}

func (pub *LogPublisher) Publish(line string, step *testv1alpha1.StepStatus) error {
	if err := pub.ensureConnected(); err != nil {
		return err
	}

	msg := &LogMessage{Line: line, Id: step.PluralId}
	return pub.Channel.Push("stdo", msg, func(payload interface{}) {})
}

func (pub *LogPublisher) Close() error {
	return pub.Channel.Leave(map[string]string{})
}

func (pub *LogPublisher) ensureConnected() error {
	if !pub.Socket.Connected {
		if err := pub.Socket.Connect(); err != nil {
			return err
		}
	}

	if pub.Open {
		return nil
	}

	channel, err := pub.Socket.Join(pub, fmt.Sprintf("tests:%s", pub.Test.Status.PluralId))
	if err != nil {
		return err
	}
	pub.Channel = channel
	pub.Open = true
	return nil
}

// phx.ChannelReceiver implementation
func (pub *LogPublisher) OnJoin(payload interface{})      {}
func (pub *LogPublisher) OnJoinError(payload interface{}) {}
func (pub *LogPublisher) OnChannelClose(payload interface{}) {
	pub.Open = false
}
func (pub *LogPublisher) OnMessage(event string, payload interface{})                             {}
func (pub *LogPublisher) OnMessageToReply(message interface{}, event string, payload interface{}) {}
