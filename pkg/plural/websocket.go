package plural

import (
	"fmt"
	phx "github.com/Douvi/gophoenix"
	"net/http"
	"net/url"
	"sync"
)

type Socket struct {
	mu        sync.Mutex
	Client    *phx.Client
	Config    *Config
	Connected bool
}

func (s *Socket) NotifyDisconnect() {
	s.Connected = false
}

func (s *Socket) NotifyConnect() {
	s.Connected = true
}

func WebSocket(config *Config) (socket Socket) {
	socket.Config = config
	socket.Client = phx.NewWebsocketClient(&socket)
	return
}

func (socket *Socket) Connect() error {
	socket.mu.Lock()
	defer socket.mu.Unlock()
	if socket.Connected {
		return nil
	}

	conf := socket.Config
	url, err := url.Parse(fmt.Sprintf("wss://%s/socket/websocket?token=%s", conf.PluralEndpoint(), conf.Token))
	if err != nil {
		return err
	}

	if err := socket.Client.Connect(*url, http.Header{}); err != nil {
		return err
	}

	socket.Connected = true
	return nil
}

func (socket *Socket) Join(callback phx.ChannelReceiver, topic string) (*phx.Channel, error) {
	return socket.Client.Join(callback, topic, map[string]string{})
}
