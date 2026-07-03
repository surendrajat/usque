//go:build !linux && !windows && !darwin

package nativetun

import (
	"errors"
	"io"

	"github.com/Diniboy1123/usque/api"
)

func createDevice(opts Options) (api.TunnelDevice, string, io.Closer, error) {
	return nil, "", nil, errors.New("nativetun is not supported on this platform")
}
