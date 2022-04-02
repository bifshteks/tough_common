package source

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type TCP struct {
	url    string
	conn   *net.TCPConn
	reader chan []byte
}

func NewTCP(url string) *TCP {
	return &TCP{
		url:    url,
		conn:   nil,
		reader: make(chan []byte),
	}
}

func (tcp *TCP) GetUrl() string {
	return tcp.url
}

func (tcp *TCP) GetReader() chan []byte {
	return tcp.reader
}

func (tcp *TCP) Connect(ctx context.Context) error {
	logrus.Debugf("tcp.Connect() on %s", tcp.url)
	select {
	case <-ctx.Done():
		return nil
	default:
		logrus.Infof("Connecting to VNC on %s", tcp.url)
		conn, err := net.DialTimeout("tcp", tcp.url, 3*time.Second)
		if err != nil {
			logrus.Errorln("VNC dial failed:", err)
			return err
		}
		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			panic("cannot convert to tcpConn")
		}
		tcp.conn = tcpConn
		logrus.Infof("Connected to vnc on %s", tcp.url)
		go func() {
			<-ctx.Done()
			tcp.Close()
		}()
		return nil
	}
}

func (tcp *TCP) Consume(ctx context.Context) error {
	defer logrus.Debugln("tcp.Start() ends")
	logrus.Debugln("tcp.Start() call")

	// don't need to catch context done - we already created a goroutine in .Connect() method
	// waiting for that
	for {
		buffer := make([]byte, 1024)
		n, err := tcp.conn.Read(buffer)
		if err != nil {
			netErr, ok := err.(net.Error)
			temporary := ok && netErr.Temporary()
			if temporary {
				continue
			}
			return errors.New(fmt.Sprintf(
				"Could not read from tcp on %s: %s", tcp.url, err,
			))
		}
		tcp.reader <- buffer[:n]
	}
}

func (tcp *TCP) Write(msg []byte) error {
	_, err := tcp.conn.Write(msg)
	return err
}

func (tcp *TCP) Close() {
	defer logrus.Debugln("tcp.Close() ends")
	logrus.Debugln("tcp.Close() call")
	err := tcp.conn.Close()
	if err != nil {
		logrus.Errorln("Could not close connection to tcp:", err)
	}
	tcp.conn = nil
	close(tcp.reader)
}
