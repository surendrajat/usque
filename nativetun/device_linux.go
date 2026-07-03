//go:build linux

package nativetun

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/Diniboy1123/usque/api"
	"github.com/Diniboy1123/usque/config"
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

func createDevice(opts Options) (api.TunnelDevice, string, io.Closer, error) {
	dev, err := water.New(water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name:    opts.Name,
			Persist: opts.Persist,
		},
	})
	if err != nil {
		return nil, "", nil, err
	}

	name := dev.Name()

	if opts.Iproute2 {
		link, err := netlink.LinkByName(name)
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to get link: %v", err)
		}

		if err := netlink.LinkSetMTU(link, opts.MTU); err != nil {
			return nil, "", nil, fmt.Errorf("failed to set MTU: %v", err)
		}
		if opts.TunnelIPv4 {
			if err := netlink.AddrAdd(link, &netlink.Addr{
				IPNet: &net.IPNet{
					IP:   net.ParseIP(config.AppConfig.IPv4),
					Mask: net.CIDRMask(32, 32),
				}}); err != nil {
				return nil, "", nil, fmt.Errorf("failed to add IPv4 address: %v", err)
			}
		}
		if opts.TunnelIPv6 {
			if err := netlink.AddrAdd(link, &netlink.Addr{
				IPNet: &net.IPNet{
					IP:   net.ParseIP(config.AppConfig.IPv6),
					Mask: net.CIDRMask(128, 128),
				}}); err != nil {
				return nil, "", nil, fmt.Errorf("failed to add IPv6 address: %v", err)
			}
		}
		if err := netlink.LinkSetUp(link); err != nil {
			return nil, "", nil, fmt.Errorf("failed to set link up: %v", err)
		}
	} else {
		log.Println("Skipping IP address and link setup. You should set the link up manually.")
		log.Println("Config has the following IP addresses:")
		log.Printf("IPv4: %s", config.AppConfig.IPv4)
		log.Printf("IPv6: %s", config.AppConfig.IPv6)
	}

	return api.NewWaterAdapter(dev), name, dev, nil
}
