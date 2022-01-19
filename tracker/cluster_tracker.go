package tracker

import (
	"fmt"
	ttlcache "github.com/ReneKroon/ttlcache/v2"
	"github.com/lonelycode/roomystan/config"
	"sort"
	"time"
)

type DeviceData struct {
	Name       string
	Distance   float64
	LastUpdate time.Time
}

type Member struct {
	ID      string
	Devices *ttlcache.Cache
}

type Cluster struct {
	Members *ttlcache.Cache
}

func NewCluster() *Cluster {
	c := &Cluster{
		Members: ttlcache.NewCache(),
	}

	c.Members.SetTTL(time.Duration(15 * time.Minute))
	c.Members.SkipTTLExtensionOnHit(true)

	return c
}

func (c *Cluster) UpdateMember(memberID, deviceID string, distance float64) {
	var mem = NewMember(memberID, config.Get().DeviceTTL)
	var ok bool
	if val, err := c.Members.Get(memberID); err != ttlcache.ErrNotFound {
		mem, ok = val.(*Member)
		if !ok {
			panic("value is not member type!")
		}
	}

	mem.AddOrUpdateDevice(deviceID, distance)
	c.Members.Set(memberID, mem)
	fmt.Printf("updated %s in %s with %v\n", memberID, deviceID, distance)
}

type DeviceStatus struct {
	Name     string
	SeenBy   string
	Distance float64
	LastSeen time.Time
}

func (c *Cluster) EstimateDeviceLocations() map[string]string {
	iMembers := c.Members.GetItems()

	deviceStatus := make(map[string][]*DeviceStatus, 0)
	locations := map[string]string{}

	for _, iMemberData := range iMembers {
		mDat, ok := iMemberData.(*Member)
		if !ok {
			panic("invalid member type")
		}

		iDevices := mDat.Devices.GetItems()
		for _, iDeviceData := range iDevices {
			dev, ok := iDeviceData.(*DeviceData)
			if !ok {
				panic("invalid device type")
			}
			stat := &DeviceStatus{
				Name:     dev.Name,
				SeenBy:   mDat.ID,
				Distance: dev.Distance,
				LastSeen: dev.LastUpdate,
			}

			_, ok = deviceStatus[dev.Name]
			if !ok {
				deviceStatus[dev.Name] = make([]*DeviceStatus, 0)
			}

			deviceStatus[dev.Name] = append(deviceStatus[dev.Name], stat)
		}
	}

	for name, stats := range deviceStatus {
		sort.Slice(stats, func(i, j int) bool {
			return stats[i].Distance < stats[j].Distance
		})

		if len(stats) > 0 {
			locations[name] = stats[0].SeenBy
		}
	}

	return locations
}

func NewMember(name string, ttl int) *Member {
	m := &Member{
		ID:      name,
		Devices: ttlcache.NewCache(),
	}

	m.Devices.SetTTL(time.Duration(ttl) * time.Second)
	m.Devices.SkipTTLExtensionOnHit(true)

	return m
}

func (m *Member) AddOrUpdateDevice(name string, distance float64) {
	d := &DeviceData{
		LastUpdate: time.Now(),
		Name:       name,
		Distance:   distance,
	}

	m.Devices.Set(name, d)
}
