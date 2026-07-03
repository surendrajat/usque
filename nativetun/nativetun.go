// Package nativetun runs Warp as a native TUN device: it creates the platform TUN,
// then maintains the MASQUE tunnel until the context is cancelled. The nativetun CLI
// command and any program embedding usque as a library both call Run, so the tunnel
// code path is identical on every OS.
package nativetun

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Diniboy1123/usque/api"
	"github.com/Diniboy1123/usque/config"
	"github.com/Diniboy1123/usque/internal"
)

// Options configures a native TUN tunnel. Fields mirror the nativetun CLI flags.
type Options struct {
	Name              string        // interface name ("" = platform default)
	MTU               int
	KeepalivePeriod   time.Duration
	InitialPacketSize uint16
	ReconnectDelay    time.Duration
	AlwaysReconnect   bool
	ConnectPort       int
	UseHTTP2          bool
	UseIPv6Endpoint   bool // select an IPv6 MASQUE endpoint
	TunnelIPv4        bool // configure the device's IPv4 address
	TunnelIPv6        bool // configure the device's IPv6 address
	Iproute2          bool // linux: set addresses + bring the link up
	Persist           bool // linux: keep the interface after exit
	SNI               string
	Insecure          bool
	OnConnect         string // exec hook path (fire-and-forget)
	OnDisconnect      string
	OnConnectFunc     func() // in-process connect callback (library embedders)
	OnDisconnectFunc  func()
	// OnDeviceUp, if set, is called with the resolved interface name right after the
	// device is created (before the first connect) so an embedder can set up routing.
	OnDeviceUp func(iface string)
}

// Run creates the native TUN device and maintains the MASQUE tunnel until ctx is
// cancelled. It blocks (like the CLI's select{}); embedders should call it in a
// goroutine and cancel ctx to stop.
func Run(ctx context.Context, opts Options) error {
	if !config.ConfigLoaded {
		return errors.New("config not loaded; register first")
	}
	privKey, err := config.AppConfig.GetEcPrivateKey()
	if err != nil {
		return err
	}
	peerPubKey, err := config.AppConfig.GetEcEndpointPublicKey()
	if err != nil {
		return err
	}
	cert, err := internal.GenerateCert(privKey, &privKey.PublicKey)
	if err != nil {
		return err
	}
	sni := opts.SNI
	if sni == "" {
		sni = internal.ConnectSNI
	}
	if opts.Insecure {
		config.WarnInsecure()
	}
	tlsConfig, err := api.PrepareTlsConfig(privKey, peerPubKey, cert, sni, opts.Insecure)
	if err != nil {
		return err
	}
	endpoint, err := config.SelectEndpointFromConfig(opts.UseHTTP2, opts.UseIPv6Endpoint, opts.ConnectPort)
	if err != nil {
		return err
	}
	if opts.UseHTTP2 {
		config.LogHTTP2Endpoint(endpoint)
	}

	dev, name, closer, err := createDevice(opts)
	if err != nil {
		return err
	}
	if closer != nil {
		defer closer.Close()
	}
	log.Printf("Created TUN device: %s", name)

	if opts.OnDeviceUp != nil {
		opts.OnDeviceUp(name)
	}

	log.Println("Tunnel established, you may now set up routing and DNS")
	api.MaintainTunnel(ctx, api.MaintainTunnelConfig{
		TLSConfig:         tlsConfig,
		KeepalivePeriod:   opts.KeepalivePeriod,
		InitialPacketSize: opts.InitialPacketSize,
		Endpoint:          endpoint,
		Device:            dev,
		MTU:               opts.MTU,
		ReconnectDelay:    opts.ReconnectDelay,
		AlwaysReconnect:   opts.AlwaysReconnect,
		UseHTTP2:          opts.UseHTTP2,
		OnConnect:         opts.OnConnect,
		OnDisconnect:      opts.OnDisconnect,
		OnConnectFunc:     opts.OnConnectFunc,
		OnDisconnectFunc:  opts.OnDisconnectFunc,
		HookEnv: map[string]string{
			"USQUE_MODE":  "nativetun",
			"USQUE_IFACE": name,
			"USQUE_IPV4":  config.AppConfig.IPv4,
			"USQUE_IPV6":  config.AppConfig.IPv6,
		},
	})
	return nil
}
