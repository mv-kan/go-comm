package gocomm

import (
	"fmt"
	"time"

	"github.com/tarm/serial"
)

type Port interface {
	Close() error
	Flush() error
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Reopen() error
}

func NewPort(devPath string, baud int, readTimeout time.Duration) (Port, error) {
	p := &port{
		devPath:     devPath,
		baud:        baud,
		readTimeout: readTimeout,
	}
	err := p.Reopen()
	return p, err
}

type port struct {
	devPath     string
	readTimeout time.Duration
	baud        int
	p           *serial.Port
}

func (p *port) Close() error {
	if p.p == nil {
		return ErrNotConnected
	}
	return p.p.Close()
}
func (p *port) Flush() error {
	if p.p == nil {
		return ErrNotConnected
	}
	return p.p.Flush()
}
func (p *port) Read(b []byte) (n int, err error) {
	if p.p == nil {
		return 0, ErrNotConnected
	}
	return p.p.Read(b)
}
func (p *port) Write(b []byte) (n int, err error) {
	if p.p == nil {
		return 0, ErrNotConnected
	}
	return p.p.Write(b)
}
func (p *port) Reopen() error {
	if p.p != nil {
		p.p.Close()
	}
	sc := &serial.Config{
		Name:        p.devPath,
		Baud:        p.baud,
		ReadTimeout: p.readTimeout,
	}
	port, err := serial.OpenPort(sc)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDevice, err)
	}
	p.p = port
	return nil
}
