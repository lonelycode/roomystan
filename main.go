package main

import (
	"fmt"
	"github.com/lonelycode/roomystan/scanner"
	"github.com/lonelycode/roomystan/tracker"
	"time"
)

func start() {
	b := scanner.New()
	t := tracker.New("local", []string{"Pam", "room-assistant companion"}, 3)
	c := tracker.NewCluster()
	t.OnUpdate = func(cluster *tracker.Cluster) func(trackerID string, deviceID string, distance float64) {
		return func(trackerID string, deviceID string, distance float64) {
			cluster.UpdateMember(trackerID, deviceID, distance)
		}
	}(c)

	go func(cluster *tracker.Cluster) {
		for {
			time.Sleep(10 * time.Second)
			fmt.Println(cluster.EstimateDeviceLocations())
		}
	}(c)

	b.Scan(t.Update)
}

func main() {
	start()
}
