package gocomm

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

type Connection interface {
	Close() error
}
type conn struct {
	p                 Port
	isClosing         bool
	reconnMu          sync.Mutex
	reconnRequestChan chan time.Time
	done              chan struct{}
}
type Message struct {
	Data string
	Err  error
}

func NewConnection(
	p Port,
	input <-chan string,
	writeInterval time.Duration,
	delimRead string,
	delimWrite string,
) (Connection, <-chan Message, error) {
	if err := p.Reopen(); err != nil {
		return nil, nil, err
	}
	c := conn{
		p:    p,
		done: make(chan struct{}),
	}
	msgChan := make(chan Message)
	go c.reconnecting(msgChan)
	go c.reading(msgChan, delimRead)
	go c.writing(input, msgChan, writeInterval, delimWrite)
	return &c, msgChan, nil
}
func (c *conn) Close() error {
	c.reconnMu.Lock()
	defer c.reconnMu.Unlock()
	c.isClosing = true
	close(c.done)
	return c.p.Close()
}
func (c *conn) reconnecting(
	msgChan chan<- Message,
) {
	lastReconnect := time.Time{}
	for {
		select {
		case <-c.done:
			return
		case req := <-c.reconnRequestChan:
			if req.Before(lastReconnect) {
				log.Printf("reconnecting: req=%v, lastReconnect=%v, skipping reconn req", req, lastReconnect)
				c.reconnMu.Unlock()
				continue
			}
			log.Printf("reconnecting: start reconnecting")
			log.Printf("reconnecting inner: before u.p.Close()")
			c.p.Close()
			log.Printf("reconnecting inner: for loop begin")
			timer := time.NewTimer(0)
		out:
			for {
				log.Printf("reconnecting inner: for")
				select {
				case <-c.done:
					break out
				case <-timer.C:
					log.Printf("reconnecting inner: case <-timer.C")
					err := c.p.Reopen()
					if err == nil {
						log.Printf("reconnecting: connected!")
						break out
					} else {
						log.Printf("reconnecting: err=%v", err)
						msgChan <- Message{Err: err}
						timer.Reset(time.Millisecond * 100)
					}
				}
			}
			log.Printf("reconnecting: mu.Unlock()")
			lastReconnect = time.Now()
			c.reconnMu.Unlock()
		}
	}
}

func (c *conn) reading(
	msgChan chan<- Message,
	delimRead string,
) {
	buf := make([]byte, 1)
	line := make([]byte, 4096)
	n := 0
	for {
		select {
		case <-c.done:
			return
		default:
			// DO NOT USE MUTEX HERE
			_, err := c.p.Read(buf)
			if err != nil {
				log.Printf("reading: time=%v communication error, err=%v", time.Now().UnixMilli(), err)
				c.sendReconnReq(msgChan, err)
				continue
			}

			readChar := buf[0]
			// log.Printf("reading: time=%v %v", time.Now().UnixMilli(), string(buf[0]))

			if string(readChar) == delimRead || n >= 4096 {
				//s.logger.Debug("read line: ", string(line))
				log.Printf("reading: time=%v <%s>", time.Now().UnixMilli(), string(line[:n]))
				msgChan <- Message{Data: string(line[:n])}
				n = 0
			} else {
				line[n] = readChar
				n++
			}
		}
	}
}

func (c *conn) writing(
	input <-chan string,
	msgChan chan<- Message,
	writeInterval time.Duration,
	delimWrite string,
) {
	timeWait := writeInterval
	var lastSend time.Time
	for {
		select {
		case <-c.done:
			return
		case str := <-input:
			log.Printf("writing: accept <%s>, lastSeen: %v ago", strings.TrimSpace(str), time.Since(lastSend))
			if time.Since(lastSend) < timeWait {
				timeToWait := timeWait - time.Since(lastSend)
				log.Printf("writing: <%s> waiting for %v", str, timeToWait)
				time.Sleep(timeToWait)
			}
			_, err := c.p.Write([]byte(fmt.Sprintf("%s%s", str, delimWrite)))
			if err != nil {
				log.Printf("writing: <%s> time=%v communication error, err=%v", str, time.Now().UnixMilli(), err)
				c.sendReconnReq(msgChan, err)
				continue
			} else {
				lastSend = time.Now()
			}
		}
	}
}

func (c *conn) sendReconnReq(msgChan chan<- Message, err error) {
	if c.isClosing {
		return
	}
	now := time.Now()
	log.Printf("sendReconnReq, now=%v", now)
	c.reconnMu.Lock()
	msgChan <- Message{Err: err}
	c.reconnRequestChan <- now
	log.Printf("sendReconnReq finished")
}
