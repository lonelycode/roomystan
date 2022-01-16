package scanner

import (
	"log"
	"tinygo.org/x/bluetooth"
)

type BLE struct {
	adapter *bluetooth.Adapter
}

type CallBack func(adapter *bluetooth.Adapter, device bluetooth.ScanResult)

func New() *BLE {
	return &BLE{
		adapter: bluetooth.DefaultAdapter,
	}
}

func (b *BLE) Scan(cb CallBack) {
	// Enable BLE interface.
	err := b.adapter.Enable()
	if err != nil {
		log.Fatal("failed to start adapter", err)
	}

	// Start scanning.
	println("Scanning...")
	err = b.adapter.Scan(cb)
	if err != nil {
		log.Fatal("scanning failed", err)
	}
}

func PrintCallBack(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
	println("found device:", device.Address.String(), device.RSSI, device.LocalName())
}
