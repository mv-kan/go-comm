package gocomm

import "errors"

var (
	ErrDevice       = errors.New("device error")
	ErrNotConnected = errors.New("not connected")
)
