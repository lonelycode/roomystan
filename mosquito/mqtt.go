package mosquito

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/lonelycode/roomystan/config"
	uuid "github.com/satori/go.uuid"
)

const (
	statusTopic string = "roomystan/deviceStatus"
)

type MQTTHandler struct {
	cfg            *config.MQTTConf
	client         mqtt.Client
	PayloadHandler func(client mqtt.Client, msg mqtt.Message)
}

type Payload struct {
	Device   string
	Distance float64
	Member   string
}

func New(cfg *config.MQTTConf) *MQTTHandler {
	return &MQTTHandler{cfg: cfg}
}

func (m *MQTTHandler) Connect() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", m.cfg.Address, m.cfg.Port))
	opts.SetClientID(fmt.Sprintf("roomystan-mqtt-%s", uuid.NewV4().String())) // need unique IDs for each client
	opts.SetUsername(m.cfg.User)
	opts.SetPassword(m.cfg.Pass)

	c := mqtt.NewClient(opts)

	token := c.Connect()
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	m.client = c

	return nil
}

func (m *MQTTHandler) messageHandler(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

func (m *MQTTHandler) ListenForClusterUpdates() {
	if m.PayloadHandler == nil {
		m.PayloadHandler = m.messageHandler
	}

	token := m.client.Subscribe(statusTopic, 1, m.PayloadHandler)
	token.Wait()
}

func (m *MQTTHandler) SendClusterUpdate(data []byte) {
	token := m.client.Publish(statusTopic, 0, false, data)
	token.Wait()
}
