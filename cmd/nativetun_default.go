//go:build !linux && !windows && !darwin

package cmd

var longDescription = "Expose Warp as a native TUN device that accepts any IP traffic." +
	" This command is not supported on your platform."
