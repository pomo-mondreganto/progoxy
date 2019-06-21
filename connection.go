package main

import (
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

type Connection struct {
	net.Conn
	IdleTimeout  time.Duration
	MaxReadBytes int64

	OutConn net.Conn
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
	return
}

func (c *Connection) StartReadingSrc(outCh chan []byte, errCh chan error) {
	buf := make([]byte, 1024)
	for {
		curPos := 0
		for {
			cnt, err := c.Read(buf[curPos:])
			if cnt > 0 {
				outCh <- buf[curPos : curPos+cnt]
				curPos += cnt
			}
			if err != nil {
				if err != io.EOF {
					errCh <- err
				}
				return
			}
			if curPos >= len(buf) {
				break
			}
		}
	}
}

func (c *Connection) StartReadingDst(outCh chan []byte, errCh chan error) {
	buf := make([]byte, 1024)
	for {
		curPos := 0
		for {
			cnt, err := c.OutConn.Read(buf[curPos:])
			if cnt > 0 {
				outCh <- buf[curPos : curPos+cnt]
				curPos += cnt
			}
			if err != nil {
				if err != io.EOF {
					errCh <- err
				}
				return
			}
			if curPos >= len(buf) {
				break
			}
		}
	}
}