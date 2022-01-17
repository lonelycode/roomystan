package tracker

import (
	"fmt"
	"github.com/konimarti/kalman"
	"github.com/konimarti/lti"
	"gonum.org/v1/gonum/mat"
	"math"
	"time"
	"tinygo.org/x/bluetooth"
)

const (
	MeasuredPower = -75.0
	BLE_LowPower  = 2.0
	BLE_HighPower = 4.0
)

type Tracker struct {
	ID             string
	DeviceIDs      []string
	LocalTrackData map[string]*WindowData
	PowerValue     float64
	Samples        int
	OnUpdate       func(trackerID string, deviceID string, distance float64)
}

type WindowData struct {
	Window      []float64
	SampleSet   int
	CurrentDist float64
	LastUpdated time.Time
}

func New(id string, ids []string, sample int) *Tracker {
	return &Tracker{
		ID:             id,
		DeviceIDs:      ids,
		LocalTrackData: map[string]*WindowData{},
		PowerValue:     BLE_LowPower,
		Samples:        sample,
	}
}

func PrintUpdate(trackerID string, deviceID string, distance float64) {
	fmt.Printf("Updated Distance for '%s/%s' to %vm\n", trackerID, deviceID, distance)
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
					Window: make([]float64, t.Samples),
				}
			}

			// only work with a full sample set
			setNum := t.LocalTrackData[id].SampleSet
			t.LocalTrackData[id].Window[setNum] = rssi
			if setNum < t.Samples-1 {
				t.LocalTrackData[id].SampleSet += 1
			} else {
				t.LocalTrackData[id].SampleSet = 0
			}

			// Calc only on x samples
			if t.LocalTrackData[id].SampleSet+1 == t.Samples {
				filtered := t.Filter(t.LocalTrackData[id].Window)
				distNow := t.RSSIDist(filtered)

				t.LocalTrackData[id].CurrentDist = distNow
				t.LocalTrackData[id].LastUpdated = time.Now()
				fmt.Printf("Updated Distance for '%v (%s)' (RSSI %v, Filter: %v) to %vm\n", id, name, rssi, filtered, distNow)
				if t.OnUpdate != nil {
					t.OnUpdate(t.ID, v, distNow)
				}
			}

			break
		}
	}
}

func (t *Tracker) RSSIDist(measurement float64) float64 {
	return math.Pow(10, ((MeasuredPower - measurement) / (10.0 * t.PowerValue)))
}

func (t *Tracker) Filter(row []float64) float64 {
	// define LTI system
	lti := lti.Discrete{
		Ad: mat.NewDense(1, 1, []float64{1}),
		Bd: mat.NewDense(1, 1, nil),
		C:  mat.NewDense(1, 1, []float64{1}),
		D:  mat.NewDense(1, 1, nil),
	}

	// system noise / process model covariance matrix ("Systemrauschen")
	Gd := mat.NewDense(1, 1, []float64{1})

	ctx := kalman.Context{
		// initial state
		X: mat.NewVecDense(1, []float64{row[0]}),
		// initial covariance matrix
		P: mat.NewDense(1, 1, []float64{0}),
	}

	// create ROSE filter
	gammaR := 9.0
	alphaR := 0.5
	alphaM := 0.3
	filter := kalman.NewRoseFilter(lti, Gd, gammaR, alphaR, alphaM)

	// no control
	u := mat.NewVecDense(1, nil)

	tot := 0.0
	for _, c := range row {
		// new measurement
		y := mat.NewVecDense(1, []float64{c})

		// apply filter
		filter.Apply(&ctx, y, u)

		// get corrected state vector
		state := filter.State()

		// print out input and output signals
		//fmt.Printf("%3.8f,%3.8f\n", y.AtVec(0), state.AtVec(0))
		tot += state.AtVec(0)
	}

	return tot / float64(t.Samples)
}
