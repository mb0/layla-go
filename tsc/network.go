package tsc

import (
	"net"
	"time"
)

// Dial opens and returns a network printer connection or returns an error.
func Dial(addr string) (*Conn, error) { return DialTimeout(addr, DefaultTimeout) }

func DialTimeout(addr string, timeout time.Duration) (*Conn, error) {
	a, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	if a.Port == 0 {
		a.Port = DefaultPort
	}
	c, err := net.DialTCP("tcp", nil, a)
	if err != nil {
		return nil, err
	}
	return &Conn{addr, c, timeout}, nil
}

// DiscoverNet discovers and returns a list of checked local network device infos or an error.
func DiscoverNet(timeout time.Duration) ([]NetInfo, error) {
	d := NewNetDiscovery(timeout)
	var infos []NetInfo
	err := d.Run(func(info NetInfo) bool {
		infos = append(infos, info)
		return true
	})
	return infos, err
}

type NetInfo struct {
	IP     string
	Mac    string
	Name   string
	Status Status
}

func parseNetInfo(msg []byte) NetInfo {
	return NetInfo{
		IP:     net.IP(msg[44:48]).String(),
		Mac:    net.HardwareAddr(msg[22:28]).String(),
		Name:   string(msg[52:58]),
		Status: Status(msg[40]),
	}
}

var udpReq = []byte{
	0, 0x20, 0, 1, 0, 1, 8, 0,
	0, 2, 0, 0, 0, 1, 0, 0,
	1, 0, 0, 0, 0, 0, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff,
	0, 0, 0, 0,
}

type NetDiscovery struct {
	Port    int
	Bcast   net.IP
	Timeout time.Duration
	Log     func(error)
}

func NewNetDiscovery(timeout time.Duration) *NetDiscovery {
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	return &NetDiscovery{Port: 22368, Bcast: net.IPv4bcast, Timeout: timeout}
}

func (d *NetDiscovery) Run(handler func(NetInfo) bool) error {
	laddr := &net.UDPAddr{IP: net.IPv4zero, Port: d.Port}
	raddr := &net.UDPAddr{IP: d.Bcast, Port: d.Port}
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	msg := make([]byte, 512)
	copy(msg, udpReq)
	_, err = conn.WriteToUDP(msg[:len(udpReq)], raddr)
	if err != nil {
		return err
	}
	if d.Timeout > 0 {
		err = conn.SetReadDeadline(time.Now().Add(d.Timeout))
		if err != nil {
			return err
		}
	}
	for {
		n, _, err := conn.ReadFromUDP(msg)
		switch e := err.(type) {
		case net.Error:
			if e.Timeout() {
				return nil
			}
			if e.Temporary() {
				if d.Log != nil {
					d.Log(err)
				}
				continue
			}
			return err
		case nil:
			if n <= 58 {
				continue
			}
			info := parseNetInfo(msg)
			if info.IP == "0.0.0.0" {
				continue
			}
			if !handler(info) {
				return nil
			}
		default:
			return err
		}
	}
}
