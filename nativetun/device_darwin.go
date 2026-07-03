//go:build darwin

package nativetun

import (
	"fmt"
	"io"

	"github.com/Diniboy1123/usque/api"
	"github.com/Diniboy1123/usque/config"
	"github.com/Diniboy1123/usque/internal"
	"golang.zx2c4.com/wireguard/tun"
)

// macOS utun devices prepend a 4-byte address-family header; wireguard-go/tun
// exposes it as a read/write offset that the shared NetstackAdapter strips.
const darwinTunOffset = 4

func createDevice(opts Options) (api.TunnelDevice, string, io.Closer, error) {
	name := opts.Name
	if name == "" {
		name = "utun"
	}

	dev, err := tun.CreateTUN(name, opts.MTU)
	if err != nil {
		return nil, "", nil, err
	}

	name, err = dev.Name()
	if err != nil {
		dev.Close()
		return nil, "", nil, err
	}

	if opts.TunnelIPv4 {
		if err := internal.SetIPv4Address(name, config.AppConfig.IPv4, "255.255.255.255"); err != nil {
			dev.Close()
			return nil, "", nil, fmt.Errorf("failed to set IPv4 address: %v", err)
		}
		if err := internal.SetIPv4MTU(name, opts.MTU); err != nil {
			dev.Close()
			return nil, "", nil, fmt.Errorf("failed to set IPv4 MTU: %v", err)
		}
	}

	if opts.TunnelIPv6 {
		if err := internal.SetIPv6Address(name, config.AppConfig.IPv6, "128"); err != nil {
			dev.Close()
			return nil, "", nil, fmt.Errorf("failed to set IPv6 address: %v", err)
		}
		if err := internal.SetIPv6MTU(name, opts.MTU); err != nil {
			dev.Close()
			return nil, "", nil, fmt.Errorf("failed to set IPv6 MTU: %v", err)
		}
	}

	if err := internal.SetInterfaceUp(name); err != nil {
		dev.Close()
		return nil, "", nil, fmt.Errorf("failed to bring interface up: %v", err)
	}

	return api.NewNetstackAdapterWithOffset(dev, darwinTunOffset), name, dev, nil
}
