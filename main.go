package main

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/lonelycode/roomystan/config"
	"github.com/lonelycode/roomystan/leader"
	"github.com/lonelycode/roomystan/mosquito"
	"github.com/lonelycode/roomystan/scanner"
	"github.com/lonelycode/roomystan/tracker"
	"time"
)

func start(devices []string) {
	b := scanner.New()
	c := tracker.NewCluster()

	mq := mosquito.New(config.Get().MQTT)
	err := mq.Connect()
	if err != nil {
		panic(err)
	}
	mq.PayloadHandler = OnRecieveClusterUpdateFunc(c)
	mq.ListenForClusterUpdates()

	stopChan := make(chan struct{})
	l := leader.Leader{
		OnLeader: func() error {
			go broadcastDeviceLocations(stopChan, c, mq)
			return nil
		},
		OnFollower: func() error {
			select {
			case stopChan <- struct{}{}:
				//
			default:
				//
			}
			return nil
		},
	}

	go l.Start(config.Get().Cluster.Members, config.Get().Cluster.ListenOn)

	t := tracker.New(config.Get().Name, devices, 3)
	t.OnUpdate = func(cluster *tracker.Cluster) tracker.CallBackFunc {
		return func(trackerID string, deviceID string, distance float64) {
			data, err := json.Marshal(&mosquito.Payload{
				Device:   deviceID,
				Distance: distance,
				Member:   trackerID,
			})

			if err != nil {
				panic(err)
			}

			mq.SendClusterUpdate(data)
		}
	}(c)

	b.Scan(t.Update)
}

type HassSensorPayload struct {
	Devices []*DeviceInfo
}

type DeviceInfo struct {
	Name     string
	Location string
}

func hassPayloadFromCluster(dat map[string]string) *HassSensorPayload {
	pl := &HassSensorPayload{Devices: make([]*DeviceInfo, 0)}
	for k, v := range dat {
		di := &DeviceInfo{
			Name:     k,
			Location: v,
		}
		pl.Devices = append(pl.Devices, di)
	}

	return pl
}

func broadcastDeviceLocations(stop chan struct{}, cluster *tracker.Cluster, m *mosquito.MQTTHandler) {
	ticker := time.NewTicker(5 * time.Second)
	fmt.Println("starting leader broadcast")
	for {
		select {
		case <-stop:
			fmt.Println("stopping leader broadcast")
			return
		case <-ticker.C:
			//fmt.Println(cluster.EstimateDeviceLocations())
			b, err := json.Marshal(hassPayloadFromCluster(cluster.EstimateDeviceLocations()))
			if err != nil {
				panic(err)
			}

			m.BroadcastDeviceLocations(b)
		}
	}

}

func OnRecieveClusterUpdateFunc(cluster *tracker.Cluster) func(client mqtt.Client, msg mqtt.Message) {
	return func(client mqtt.Client, msg mqtt.Message) {
		pl := mosquito.Payload{}
		err := json.Unmarshal(msg.Payload(), &pl)
		if err != nil {
			panic(err)
		}

		cluster.UpdateMember(pl.Member, pl.Device, pl.Distance)
	}
}

func main() {
	start(config.Get().Devices)
}
