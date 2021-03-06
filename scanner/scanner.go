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
	println("scanning for BLE devices...")
	err = b.adapter.Scan(cb)
	if err != nil {
		log.Fatal("scanning failed", err)
	}
}

func PrintCallBack(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
	println("found device:", result.Address.String(), result.RSSI, result.LocalName())
}
