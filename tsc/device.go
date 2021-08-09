package tsc

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// Open opens the devices file dev and returns a new connection or an error.
func Open(dev string) (*Conn, error) {
	f, err := os.OpenFile(dev, os.O_RDWR, 660)
	if err != nil {
		return nil, err
	}
	return &Conn{dev, f, DefaultTimeout}, nil
}

// Discover discovers and returns usb devices paths that pass a status check.
// For now it only checks /dev/usb/libN on linux machines.
func DiscoverDev() (devs []string, err error) {
	switch runtime.GOOS {
	case "linux":
		search := "/dev/usb"
		du, err := os.Open(search)
		if err != nil {
			return nil, err
		}
		names, err := du.Readdirnames(0)
		du.Close()
		if err != nil {
			return nil, err
		}
		sort.Strings(names)
		for _, n := range names {
			if !strings.HasPrefix(n, "lp") {
				continue
			}
			dev := filepath.Join(search, n)
			_, err := os.Stat(dev)
			if err != nil {
				continue
			}
			devs = append(devs, dev)
		}
		return devs, nil
	default:
		return nil, fmt.Errorf("printer device discovery only implemented on linux")
	}
}
