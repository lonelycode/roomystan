package main

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/lonelycode/roomystan/config"
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

	// for debuging scanning
	go func(cluster *tracker.Cluster) {
		for {
			time.Sleep(10 * time.Second)
			fmt.Println(cluster.EstimateDeviceLocations())
		}
	}(c)

	b.Scan(t.Update)
}

func OnRecieveClusterUpdateFunc(cluster *tracker.Cluster) func(client mqtt.Client, msg mqtt.Message) {
	return func(client mqtt.Client, msg mqtt.Message) {
		fmt.Println("got message")
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
