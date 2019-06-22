package main

import (
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

func ReadToChannel(outCh chan []byte, errCh chan error, r io.Reader) {
	buf := make([]byte, 1024)
	for {
		curPos := 0
		for {
			cnt, err := r.Read(buf[curPos:])
			if cnt > 0 {
				outCh <- buf[curPos : curPos+cnt]
				curPos += cnt
			}
			if err != nil {
				errCh <- err
				return
			}
			if curPos >= len(buf) {
				break
			}
		}
	}
}

type Connection struct {
	net.Conn
	IdleTimeout  time.Duration
	MaxReadBytes int64

	OutConn io.ReadWriteCloser
}

func (c *Connection) updateDeadline() {
	newDeadline := time.Now().Add(c.IdleTimeout)
	err := c.Conn.SetDeadline(newDeadline)
	if err != nil {
		logrus.Error("Error setting connection deadline: ", err)
	}
}

func (c *Connection) Write(p []byte) (int, error) {
	c.updateDeadline()
	return c.Conn.Write(p)
}

func (c *Connection) Read(b []byte) (n int, err error) {
	c.updateDeadline()
	r := io.LimitReader(c.Conn, c.MaxReadBytes)
	n, err = r.Read(b)
	return
}

func (c *Connection) Close() (err error) {
	err = c.Conn.Close()
	if err != nil {
		_ = c.OutConn.Close()
		return
	}
	err = c.OutConn.Close()
	return
}

func (c *Connection) StartReadingSrc(outCh chan []byte, errCh chan error) {
	ReadToChannel(outCh, errCh, c)
}

func (c *Connection) StartReadingDst(outCh chan []byte, errCh chan error) {
	ReadToChannel(outCh, errCh, c.OutConn)
}
