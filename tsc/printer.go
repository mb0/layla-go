// Package tsc provides helper functions to work with TSC label printers.
package tsc

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

const (
	DefaultPort     = 9100
	DefaultReadSize = 1024
	DefaultTimeout  = 250 * time.Millisecond
)

func Auto(dev string, timeout time.Duration) (c *Conn, err error) {
	switch dev {
	case "":
	default:
		if dev[0] == '/' {
			c, err = tryConn(Open(dev))
		} else {
			if strings.IndexByte(dev, ':') < 0 {
				dev += ":0"
			}
			c, err = tryConn(Dial(dev))
		}
		if err != nil {
			log.Printf("failed to connect to %s: %v", dev, err)
			break
		}
		return c, nil
	}
	devs, _ := DiscoverDev()
	for _, dev := range devs {
		c, err := tryConn(Open(dev))
		if err != nil {
			continue
		}
		return c, nil
	}
	nfos, _ := DiscoverNet(timeout)
	for _, nfo := range nfos {
		c, err := tryConn(Dial(nfo.IP + ":0"))
		if err != nil {
			continue
		}
		return c, nil
	}
	return nil, fmt.Errorf("no tsc printer found")
}

// PrintDurations estimates the time it takes to print a number of labels
func PrintDuration(count, dots, dpi, speed int) time.Duration {
	// calculate total dots, add generously for gap and startup
	total := count * (dots + 20)
	millis := total * 10 / (dpi * speed / 100)
	return time.Duration(millis) * time.Millisecond
}

// Conn represents a connection to a TSC printer either using a device file or net connection.
type Conn struct {
	Dev     string
	Conn    ReadWriteCloser
	Timeout time.Duration
}

// Model requests and returns the printer model name or an error.
func (p *Conn) Model() (string, error) {
	err := p.Send(CmdModel)
	if err != nil {
		return "", err
	}
	res, err := p.Recv()
	return string(res), err
}

// Status requests and returns the printer status or an error.
func (p *Conn) Status() (Status, error) {
	err := p.Send(CmdStatus)
	if err != nil {
		return 0, err
	}
	res, err := p.RecvN(1)
	if err != nil {
		return 0, err
	}
	if len(res) < 1 {
		return 0, io.EOF
	}
	return Status(res[0]), nil
}

// WaitReady waits until the printer is ready or has a blocking status and sleeps for poll duration.
func (p *Conn) WaitReady(poll time.Duration) error {
	st, err := p.Status()
	if err != nil {
		return err
	}
	for st != StatusReady {
		// sleep a second and try again
		time.Sleep(poll)
		st, err = p.Status()
		if err != nil {
			return err
		}
		// error on blocking event
		// ignore head opened
		if st > StatusOpened && st < StatusPaused || st == StatusOther {
			return fmt.Errorf("printer error status: %s", st)
		}
	}
	return nil
}

// Send sends a message to the server or returns an error.
func (p *Conn) Send(cmd string) error { _, err := io.WriteString(p.Conn, cmd); return err }

// Recv receives a message with default read size and returns it or an error.
func (p *Conn) Recv() ([]byte, error) { return p.RecvN(DefaultReadSize) }

// RecvN receives a message with read size and returns it or an error.
func (p *Conn) RecvN(readSize int) ([]byte, error) {
	buf, nn := make([]byte, readSize), 0
	err := p.Conn.SetReadDeadline(time.Now().Add(p.Timeout))
	for n := 0; err == nil && nn < readSize; nn += n {
		n, err = p.Conn.Read(buf[nn:])
	}
	p.Conn.SetReadDeadline(time.Time{})
	if os.IsTimeout(err) && nn > 0 {
		err = nil
	}
	return buf[:nn], err
}

// Req sends cmd and reads the result
func (p *Conn) Req(cmd string) ([]byte, error) {
	err := p.Send(cmd)
	if err != nil {
		return nil, err
	}
	return p.Recv()
}

// Close closes the underlying connection
func (p *Conn) Close() error { return p.Conn.Close() }

// ReadWriteCloser is the abstraction covering both printer device files and net connections.
type ReadWriteCloser interface {
	io.Reader
	io.WriteCloser
	SetReadDeadline(t time.Time) error
}

func tryConn(c *Conn, err error) (*Conn, error) {
	if err != nil {
		return nil, err
	}
	_, err = c.Status()
	if err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}
