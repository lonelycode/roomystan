package tracker

import (
	"fmt"
	"github.com/asecurityteam/rolling"
	"math"
	"tinygo.org/x/bluetooth"
)

const (
	MeasuredPower = -65.0
	BLE_LowPower  = 2.0
	BLE_HighPower = 4.0
)

type Tracker struct {
	DeviceIDs      []string
	LocalTrackData map[string]*WindowData
	PowerValue     float64
	Samples        int
}

type WindowData struct {
	Window    *rolling.PointPolicy
	SampleSet int
}

func New(ids []string) *Tracker {
	return &Tracker{
		DeviceIDs:      ids,
		LocalTrackData: map[string]*WindowData{},
		PowerValue:     BLE_LowPower,
		Samples:        3,
	}
}

func (t *Tracker) Update(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
	id := device.Address.String()
	rssi := float64(device.RSSI)
	name := device.LocalName()

	for _, v := range t.DeviceIDs {
		if v == id || v == name {
			_, ok := t.LocalTrackData[id]
			if !ok {
				t.LocalTrackData[id] = &WindowData{
					Window: rolling.NewPointPolicy(rolling.NewWindow(t.Samples)),
				}
			}

			// only work with a full sample set
			t.LocalTrackData[id].Window.Append(rssi)
			if t.LocalTrackData[id].SampleSet < t.Samples {
				t.LocalTrackData[id].SampleSet += 1
			} else {
				t.LocalTrackData[id].SampleSet = 0
			}

			fmt.Println(t.LocalTrackData[id].SampleSet)

			// Debug only on 3 samples
			if t.LocalTrackData[id].SampleSet == 3 {
				currentAvg := t.LocalTrackData[id].Window.Reduce(rolling.Avg)
				distNow := math.Pow(10, ((MeasuredPower - currentAvg) / (10.0 * t.PowerValue)))
				fmt.Printf("Updated Distance for '%v (%s)' (RSSI %v / AVG %v) to %vm\n", id, name, rssi, currentAvg, distNow)
			}

			break
		}
	}
}
