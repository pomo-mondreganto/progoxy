package main

import (
	"bufio"
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

var (
	err error
)

type ServiceServer struct {
	Srv          *Service
	IdleTimeout  time.Duration
	MaxReadBytes int64

	listener   net.Listener
	conns      map[*Connection]interface{}
	mu         sync.Mutex
	inShutdown bool
}

func (ss *ServiceServer) Init() {
	ss.conns = make(map[*Connection]interface{})
}

func (ss *ServiceServer) addConnection(c *Connection) {
	defer ss.mu.Unlock()
	ss.mu.Lock()
	ss.conns[c] = struct{}{}
}

func (ss *ServiceServer) removeConnection(c *Connection) {
	defer ss.mu.Unlock()
	ss.mu.Lock()
	delete(ss.conns, c)
}

func (ss *ServiceServer) Shutdown(ctx context.Context) error {
	ss.mu.Lock()
	ss.inShutdown = true
	ss.mu.Unlock()
	ss.listener.Close()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logrus.Infof("Waiting on %v connections", len(ss.conns))
		case <-ctx.Done():
			return errors.New("server shutdown timeout")
		}
		if len(ss.conns) == 0 {
			return nil
		}
	}
}

func (ss *ServiceServer) Serve() {
	logrus.Infof("Starting server for service %s on %s", ss.Srv.Name, ss.Srv.SrcAddr)

	listener, err := net.Listen("tcp", ss.Srv.SrcAddr)
	if err != nil {
		logrus.Fatalf("Could not start server for service %s, error: %v", ss.Srv, err)
	}
	defer listener.Close()

	ss.listener = listener

	for {
		ss.mu.Lock()
		if ss.inShutdown {
			ss.mu.Unlock()
			break
		}
		ss.mu.Unlock()

		newConn, err := listener.Accept()
		if ss.inShutdown {
			break
		}
		if err != nil {
			logrus.Error("Error accepting connection: ", err)
			continue
		}
		logrus.Debug("Accepted connection from ", newConn.RemoteAddr())

		outConn, err := net.Dial("tcp", ss.Srv.DstAddr)
		if err != nil {
			logrus.Errorf("Error opening connection for service %s: %v", ss.Srv.Name, err)
			_ = newConn.Close()
			continue
		}

		logrus.Debug("Connected to ", ss.Srv.DstAddr)

		conn := &Connection{
			Conn:         newConn,
			IdleTimeout:  ss.IdleTimeout,
			MaxReadBytes: ss.MaxReadBytes,
			OutConn:      outConn,
		}
		ss.addConnection(conn)
		conn.updateDeadline()

		go ss.Handle(conn)
	}
}

func (ss *ServiceServer) Handle(conn *Connection) {
	defer func() {
		logrus.Debug("Closing connection from ", conn.RemoteAddr())
		err = conn.Close()
		if err != nil {
			logrus.Error("Error closing connection: ", err)
		}
		ss.removeConnection(conn)
	}()

	srcW := bufio.NewWriter(conn)
	dstW := bufio.NewWriter(conn.OutConn)

	srcOutCh := make(chan []byte)
	dstOutCh := make(chan []byte)

	errCh := make(chan error)
	deadline := time.After(conn.IdleTimeout)

	srcMu := sync.Mutex{}
	dstMu := sync.Mutex{}

	for {
		go conn.StartReadingSrc(srcOutCh, errCh)
		go conn.StartReadingDst(dstOutCh, errCh)
		select {
		case <-deadline:
			return
		case newSrcData := <-srcOutCh:
			srcMu.Lock()
			for _, plugWrap := range ss.Srv.SrcPlugins {
				newSrcData = plugWrap.Run(newSrcData)
			}
			_, err = dstW.Write(newSrcData)
			if err != nil {
				logrus.Error("Error writing to dst connection: ", err)
			}
			err = dstW.Flush()
			if err != nil {
				logrus.Error("Error flushing dst connection: ", err)
			}
			deadline = time.After(conn.IdleTimeout)
			srcMu.Unlock()
		case newDstData := <-dstOutCh:
			dstMu.Lock()
			for _, plugWrap := range ss.Srv.DstPlugins {
				newDstData = plugWrap.Run(newDstData)
			}
			_, err = srcW.Write(newDstData)
			if err != nil {
				logrus.Error("Error writing to src connection: ", err)
			}
			err = srcW.Flush()
			if err != nil {
				logrus.Error("Error flushing src connection: ", err)
			}
			dstMu.Unlock()
		case err := <-errCh:
			logrus.Error("Error reading from connection: ", err)
		}
	}
}
