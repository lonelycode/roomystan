package hass

import (
	"fmt"
	"github.com/gosimple/slug"
	"github.com/lonelycode/roomystan/config"
	"github.com/lonelycode/roomystan/mosquito"
	"strings"
)

type DiscoveryPayload struct {
	Platform      string `json:"platform"`
	Name          string `json:"name"`
	DeviceClass   string `json:"device_class,omitempty"`
	StateTopic    string `json:"state_topic"`
	ValueTemplate string `json:"value_template"`
}

type SensorPayload struct {
	Devices map[string]*DeviceInfo
}

type DeviceInfo struct {
	Name     string
	Location string
}

func PayloadFromCluster(dat map[string]string) *SensorPayload {
	pl := &SensorPayload{Devices: make(map[string]*DeviceInfo, 0)}
	for k, v := range dat {
		di := &DeviceInfo{
			Name:     k,
			Location: v,
		}
		pl.Devices[k] = di
	}

	notHome := []string{}
	for _, devName := range config.Get().Devices {
		_, ok := pl.Devices[devName]
		if !ok {
			// device is not home
			notHome = append(notHome, devName)
		}
	}

	for _, d := range notHome {
		di := &DeviceInfo{
			Name:     d,
			Location: "not_home",
		}
		pl.Devices[d] = di
	}

	return pl
}

func SensorName(tracker string) string {
	return fmt.Sprintf("%s_rstn_tracker", slug.Make(strings.ToLower(tracker)))
}

func BuildConfigTopic(tracker string) string {
	return fmt.Sprintf("%s/sensor/%s/location/config", config.Get().Hass.DiscoveryPrefix, SensorName(tracker))
}

func BuildStateTopic(tracker string) string {
	return fmt.Sprintf("%s/sensor/%s/location/state", config.Get().Hass.DiscoveryPrefix, SensorName(tracker))
}

func NewDiscoveryPayload(tracker string) *DiscoveryPayload {
	d := &DiscoveryPayload{
		Platform:      "mqtt",
		Name:          fmt.Sprintf("location %s", strings.ToLower(tracker)),
		StateTopic:    BuildStateTopic(tracker),
		ValueTemplate: "{{ value_json.Location }}",
	}

	return d
}

func SetupDevices(mq *mosquito.MQTTHandler) {
	for _, devName := range config.Get().Devices {
		mq.PublishTo(BuildConfigTopic(devName), NewDiscoveryPayload(devName))
	}
}

func PublishDeviceStatus(mq *mosquito.MQTTHandler, pl *SensorPayload) {
	for _, dev := range pl.Devices {
		mq.PublishTo(BuildStateTopic(dev.Name), dev)
	}
}
