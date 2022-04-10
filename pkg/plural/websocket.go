package plural

import (
	"fmt"
	phx "github.com/Douvi/gophoenix"
	"net/http"
	"net/url"
)

type Socket struct {
	Client    *phx.Client
	Config    *Config
	Connected bool
}

type connectionReceiver struct {
	client *Socket
}

func (rc *connectionReceiver) NotifyDisconnect() {
	rc.client.Connected = false
}

func (rc *connectionReceiver) NotifyConnect() {
	rc.client.Connected = true
}

func WebSocket(config *Config) (socket Socket) {
	socket.Config = config
	receiver := &connectionReceiver{client: &socket}
	socket.Client = phx.NewWebsocketClient(receiver)
	return
}

func (socket *Socket) Connect() error {
	conf := socket.Config
	url, err := url.Parse(fmt.Sprintf("wss://%s/socket/websocket?token=%s&vsn=2.0.0", conf.PluralEndpoint(), conf.Token))
	if err != nil {
		return err
	}

	return socket.Client.Connect(*url, http.Header{})
}

func (socket *Socket) Join(callback phx.ChannelReceiver, topic string) (*phx.Channel, error) {
	return socket.Client.Join(callback, topic, map[string]string{})
}
