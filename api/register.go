package api

import (
	"encoding/base64"

	"github.com/Diniboy1123/usque/config"
	"github.com/Diniboy1123/usque/internal"
)

// RegisterIdentity registers a new WARP account, enrolls a device key, and saves the
// identity to path (non-interactive, accepts Cloudflare's TOS). It's the embeddable
// equivalent of `usque register -a`, so a program using usque as a library can mint an
// identity without shelling out to the CLI. Mirrors cmd/register.go.
func RegisterIdentity(path, deviceName string) error {
	account, err := Register(internal.DefaultModel, internal.DefaultLocale, "", true)
	if err != nil {
		return err
	}
	privKey, pubKey, err := internal.GenerateEcKeyPair()
	if err != nil {
		return err
	}
	updated, err := EnrollKey(account.ID, account.Token, pubKey, deviceName)
	if err != nil {
		return err
	}
	peer := updated.Config.Peers[0]
	config.AppConfig = config.Config{
		PrivateKey:     base64.StdEncoding.EncodeToString(privKey),
		EndpointV4:     peer.Endpoint.V4[:len(peer.Endpoint.V4)-2],    // strip :0
		EndpointV6:     peer.Endpoint.V6[1 : len(peer.Endpoint.V6)-3], // strip [ and ]:0
		EndpointH2V4:   config.DefaultEndpointH2V4,
		EndpointH2V6:   config.DefaultEndpointH2V6,
		EndpointPubKey: peer.PublicKey,
		ID:             updated.ID,
		AccessToken:    account.Token,
		IPv4:           updated.Config.Interface.Addresses.V4,
		IPv6:           updated.Config.Interface.Addresses.V6,
	}
	return config.AppConfig.SaveConfig(path)
}
