package cmd

import (
	"context"
	"log"
	"time"

	"github.com/Diniboy1123/usque/config"
	"github.com/Diniboy1123/usque/internal"
	"github.com/Diniboy1123/usque/nativetun"
	"github.com/spf13/cobra"
)

var nativeTunCmd = &cobra.Command{
	Use:   "nativetun",
	Short: "Expose Warp as a native TUN device",
	Long:  longDescription,
	Run: func(cmd *cobra.Command, args []string) {
		if !config.ConfigLoaded {
			cmd.Println("Config not loaded. Please register first.")
			return
		}

		getString := func(n string) string { v, _ := cmd.Flags().GetString(n); return v }
		getBool := func(n string) bool { v, _ := cmd.Flags().GetBool(n); return v }
		getInt := func(n string) int { v, _ := cmd.Flags().GetInt(n); return v }
		getDur := func(n string) time.Duration { v, _ := cmd.Flags().GetDuration(n); return v }
		getU16 := func(n string) uint16 { v, _ := cmd.Flags().GetUint16(n); return v }

		interfaceName := getString("interface-name")
		if interfaceName != "" {
			if err := internal.CheckIfname(interfaceName); err != nil {
				log.Printf("Invalid interface name: %v", err)
				return
			}
		}

		mtu := getInt("mtu")
		if mtu != 1280 {
			log.Println("Warning: MTU is not the default 1280. This is not supported. Packet loss and other issues may occur.")
		}

		opts := nativetun.Options{
			Name:              interfaceName,
			MTU:               mtu,
			KeepalivePeriod:   getDur("keepalive-period"),
			InitialPacketSize: getU16("initial-packet-size"),
			ReconnectDelay:    getDur("reconnect-delay"),
			AlwaysReconnect:   getBool("always-reconnect"),
			ConnectPort:       getInt("connect-port"),
			UseHTTP2:          getBool("http2"),
			UseIPv6Endpoint:   getBool("ipv6"),
			TunnelIPv4:        !getBool("no-tunnel-ipv4"),
			TunnelIPv6:        !getBool("no-tunnel-ipv6"),
			Iproute2:          !getBool("no-iproute2"),
			Persist:           getBool("persist"),
			SNI:               getString("sni-address"),
			Insecure:          getBool("insecure"),
			OnConnect:         getString("on-connect"),
			OnDisconnect:      getString("on-disconnect"),
		}

		// Runs until interrupted (MaintainTunnel blocks on a never-cancelled context).
		if err := nativetun.Run(context.Background(), opts); err != nil {
			log.Fatalf("nativetun: %v", err)
		}
	},
}

func init() {
	nativeTunCmd.Flags().IntP("connect-port", "P", 443, "Used port for MASQUE connection")
	nativeTunCmd.Flags().BoolP("ipv6", "6", false, "Use IPv6 for MASQUE connection")
	nativeTunCmd.Flags().BoolP("no-tunnel-ipv4", "F", false, "Disable IPv4 inside the MASQUE tunnel")
	nativeTunCmd.Flags().BoolP("no-tunnel-ipv6", "S", false, "Disable IPv6 inside the MASQUE tunnel")
	nativeTunCmd.Flags().StringP("sni-address", "s", internal.ConnectSNI, "SNI address to use for MASQUE connection")
	nativeTunCmd.Flags().DurationP("keepalive-period", "k", 30*time.Second, "Keepalive period for MASQUE connection")
	nativeTunCmd.Flags().IntP("mtu", "m", 1280, "MTU for MASQUE connection")
	nativeTunCmd.Flags().Uint16P("initial-packet-size", "i", 0, "Custom initial packet size for MASQUE connection (default: auto with PMTU discovery)")
	nativeTunCmd.Flags().BoolP("no-iproute2", "I", false, "Linux only: Do not set up IP addresses and do not set the link up")
	nativeTunCmd.Flags().DurationP("reconnect-delay", "r", 1*time.Second, "Delay between reconnect attempts")
	nativeTunCmd.Flags().Bool("always-reconnect", false, "Always reconnect after tunnel loss, even when idle")
	nativeTunCmd.Flags().Bool("http2", false, "Use HTTP/2 over TCP+TLS instead of HTTP/3 over QUIC."+config.EndpointHelpSuffixH2)
	nativeTunCmd.Flags().Bool("insecure", false, "Disable endpoint certificate pinning and trust any certificate")
	nativeTunCmd.Flags().StringP("interface-name", "n", "", "Custom interface name for the TUN interface")
	nativeTunCmd.Flags().Bool("persist", false, "Linux only: Keep the TUN interface after exit")
	nativeTunCmd.Flags().String("on-connect", "", "Path to an executable to run after each successful tunnel connect (no args; context via USQUE_* env vars)")
	nativeTunCmd.Flags().String("on-disconnect", "", "Path to an executable to run after each tunnel disconnect (no args; context via USQUE_* env vars)")
	rootCmd.AddCommand(nativeTunCmd)
}
