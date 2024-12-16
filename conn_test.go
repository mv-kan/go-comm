package gocomm_test

import (
	"fmt"
	"testing"

	gocomm "github.com/mv-kan/go-comm"
	"github.com/stretchr/testify/assert"
)

func newMockPort() mockPort {
	p := mockPort{
		readingChan: make(chan byte),
		writingChan: make(chan []byte),
	}
	return p
}

type mockPort struct {
	readingChan chan byte
	writingChan chan []byte
	isConnected error
}

func (p *mockPort) Close() error {
	if p.isConnected != nil {
		return p.isConnected
	}
	return nil
}
func (p *mockPort) Flush() error {
	if p.isConnected != nil {
		return p.isConnected
	}
	return nil
}
func (p *mockPort) Read(b []byte) (n int, err error) {
	if p.isConnected != nil {
		return 0, p.isConnected
	}
	for i := 0; i < len(b); i++ {
		d := <-p.readingChan
		b[i] = d
	}
	return len(b), nil
}
func (p *mockPort) Write(b []byte) (n int, err error) {
	if p.isConnected != nil {
		return 0, p.isConnected
	}
	p.writingChan <- b
	return len(b), nil
}
func (p *mockPort) Reopen() error {
	if p.isConnected != nil {
		return p.isConnected
	}
	return nil
}

func TestConnectionReading_Ok(t *testing.T) {
	p := newMockPort()
	inputChan := make(chan string)
	conn, outputChan, err := gocomm.NewConnection(&p, inputChan, 0, "\n", "\n")
	defer conn.Close()

	assert.Nil(t, err)
	testMsg := "Hello\n"
	go func() {
		for i := 0; i < len(testMsg); i++ {
			p.readingChan <- testMsg[i]
		}
	}()
	msg := <-outputChan
	assert.Equal(t, testMsg[:len(testMsg)-1], msg.Data)
}

func TestConnectionWriting_Ok(t *testing.T) {
	p := newMockPort()
	inputChan := make(chan string)
	conn, _, err := gocomm.NewConnection(&p, inputChan, 0, "\n", "\n")
	defer conn.Close()
	assert.Nil(t, err)
	testMsg := "Hello"
	go func() {
		inputChan <- testMsg
	}()
	msg := <-p.writingChan
	assert.Equal(t, fmt.Sprintf("%s\n", testMsg), string(msg))
}
