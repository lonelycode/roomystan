package mosquito

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/lonelycode/roomystan/config"
	"github.com/lonelycode/roomystan/util"
	uuid "github.com/satori/go.uuid"
	"time"
)

const (
	statusTopic    string = "roomystan/deviceStatus"
	heartBeatTopic string = "roomystan/heartbeat"
	leaderTopic    string = "roomystan/leader"
)

type MQTTHandler struct {
	cfg            *config.MQTTConf
	client         mqtt.Client
	PayloadHandler func(client mqtt.Client, msg mqtt.Message)
	me             *HeartBeat
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

func (m *MQTTHandler) ListenForHeartbeats() {
	token := m.client.Subscribe(heartBeatTopic, 1, m.HeartBeatHandler)
	token.Wait()
}

func (m *MQTTHandler) HeartBeatHandler(client mqtt.Client, msg mqtt.Message) {
	hb := &HeartBeat{}
	err := json.Unmarshal(msg.Payload(), hb)
	if err != nil {
		panic(err)
	}

	if hb.IP == m.me.IP {
		return
	}
}

func (m *MQTTHandler) SendClusterUpdate(data []byte) {
	token := m.client.Publish(statusTopic, 0, false, data)
	token.Wait()
}

func (m *MQTTHandler) BroadcastDeviceLocations(data []byte) {
	token := m.client.Publish(leaderTopic, 0, false, data)
	token.Wait()
}

type HeartBeat struct {
	Name string
	IP   string
}

func (m *MQTTHandler) StartHeartbeat() {
	hb := &HeartBeat{
		Name: config.Get().Name,
		IP:   util.GetOutboundIP().String(),
	}

	m.me = hb

	asJson, err := json.Marshal(hb)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			token := m.client.Publish(heartBeatTopic, 0, false, asJson)
			token.Wait()
			time.Sleep(10 * time.Second)
		}
	}()
}
