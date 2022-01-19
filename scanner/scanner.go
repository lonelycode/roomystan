package scanner

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/bettercap/gatt"
	"log"
	"strings"
)

type BLE struct{}

func New() *BLE {
	return &BLE{}
}

var ch = make(chan gatt.Peripheral)

func (b *BLE) Scan() {
	device, err := gatt.NewDevice([]gatt.Option{
		gatt.LnxMaxConnections(1),
		gatt.LnxDeviceID(-1, true),
	}...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}
	device.Handle(gatt.PeripheralDiscovered(onPeripheralDiscovered))
	device.Init(onStateChanged)

	select {
	case p := <-ch:
		fmt.Println("connecting")
		device.Connect(p)
		device.Handle(gatt.PeripheralDiscovered(onPeripheralDiscovered))
		fmt.Println("done")
	}

	select {}
}

func onStateChanged(device gatt.Device, s gatt.State) {
	switch s {
	case gatt.StatePoweredOn:
		fmt.Println("Scanning for iBeacon Broadcasts...")
		//arnie, _ := gatt.ParseUUID(strings.ToLower("FDA50693-A4E2-4FB1-AFCF-C6EB07647825"))
		//phone, _ := gatt.ParseUUID(strings.ToLower("68D80DB9-56AA-4F45-BB56-37BC32E4185E"))
		device.Scan([]gatt.UUID{}, true)
		return
	default:
		device.StopScanning()
	}
}

func onPeripheralDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	if a.LocalName != "" {
		fmt.Printf("name: %v id: %v services: %v sData: %v\n", a.LocalName, p.ID(), a.Services, a.ServiceData)
		if len(a.Services) > 0 {
			fmt.Println(p.Services())
			ch <- p
			//svcs, err := p.DiscoverCharacteristics(nil, a.Services[0])
			//if err != nil {
			//	fmt.Println(err)
			//}
			fmt.Println("foo")
			//fmt.Println(svcs)
		}

		// === This works for the FeasyBeacons ===
		//b, err := NewiBeacon(a.ManufacturerData)
		//if err == nil {
		//	fmt.Println("Name:", a.LocalName)
		//	fmt.Println("UUID: ", b.uuid)
		//	fmt.Println("Major: ", b.major)
		//	fmt.Println("Minor: ", b.minor)
		//	fmt.Println("RSSI: ", rssi)
		//}
	}
	//b, err := NewiBeacon(a.ManufacturerData)
	//if err == nil {
	//	fmt.Println("UUID: ", b.uuid)
	//	fmt.Println("Major: ", b.major)
	//	fmt.Println("Minor: ", b.minor)
	//	fmt.Println("RSSI: ", rssi)
	//}
}

type iBeacon struct {
	uuid  string
	major uint16
	minor uint16
}

func NewiBeacon(data []byte) (*iBeacon, error) {
	if len(data) < 25 || binary.BigEndian.Uint32(data) != 0x4c000215 {
		return nil, errors.New("Not an iBeacon")
	}
	beacon := new(iBeacon)
	beacon.uuid = strings.ToUpper(hex.EncodeToString(data[4:8]) + "-" + hex.EncodeToString(data[8:10]) + "-" + hex.EncodeToString(data[10:12]) + "-" + hex.EncodeToString(data[12:14]) + "-" + hex.EncodeToString(data[14:20]))
	beacon.major = binary.BigEndian.Uint16(data[20:22])
	beacon.minor = binary.BigEndian.Uint16(data[22:24])
	return beacon, nil
}
