package main

import (
	"fmt"
	"go.uber.org/zap"
	"io"
	"net"
	"sync"
)

const (
	PROTOCOL_VERSION = 0x05
)

type state interface {
	Receive(input []byte) ([]byte, error)
}

type proxy struct {
	state          state
	input          net.Conn
	output         net.Conn
	cfg            config
	log            *zap.Logger
	negotiation    *negotiation
	authentication *passwordAuthentication
	request        *connect
}

func Start(conn net.Conn, cfg config, logger *zap.Logger) {
	p := newProxy(conn, cfg, logger)
	p.Run()
}

func newProxy(conn net.Conn, cfg config, logger *zap.Logger) *proxy {
	proxy := &proxy{
		input:          conn,
		log:            logger,
		cfg:            cfg,
		negotiation:    NewNegotiation(cfg.Auth),
		authentication: NewPasswordAuthentication(cfg.User, cfg.Pass),
	}

	proxy.request = NewRequest( conn, proxy, logger)
	proxy.state = proxy.negotiation
	return proxy
}

func (p *proxy) Run() {
	defer p.input.Close()

	buff := make([]byte, p.cfg.MTU)
	for {
		n, err := p.input.Read(buff)
		if err != nil {
			p.log.Error(fmt.Sprintf("Error read from %v: %v",p.input.RemoteAddr().String(),  err.Error()) )
			return
		}

		resp, err := p.state.Receive(buff[:n])

		if err != nil {
			p.protocolError(resp, err)
			return
		}

		responseStatus := resp[1]
		switch p.state.(type) {
		case *negotiation:
			if responseStatus == byte(NO_AUTH) {
				p.state = p.request
			} else if responseStatus == byte(PASS_AUTH) {
				p.state = p.authentication
			} else {
				return
			}
		case *passwordAuthentication:
			if responseStatus == PASS_AUTH_SUCCESS {
				p.state = p.request
			} else {
				return
			}
		case *connect:
			if responseStatus != SUCCESS {
				return
			}
		}

		_, err = p.input.Write(resp)
		if err != nil {
			p.log.Error(err.Error())
			return
		}

		if p.output != nil{
			defer p.output.Close()
			break
		}
	}

	p.log.Info(fmt.Sprintf("Start proxing %s <-> %s",p.input.RemoteAddr().String(), p.output.RemoteAddr().String()) )
	var wg sync.WaitGroup
	wg.Add(2)

	go p.pipe(p.input, p.output)
	go p.pipe(p.output, p.input)

	wg.Wait()
}

func (p *proxy) pipe(src net.Conn, dst net.Conn) {
	buf := make([]byte, p.cfg.MTU)
	for {
		n, err := src.Read(buf)
		if err != nil{
			if err == io.EOF{
				p.log.Info(fmt.Sprintf("Connection closed %s <-> %s",src.RemoteAddr().String(), dst.RemoteAddr().String()) )
			}else{
				p.log.Error(err.Error())
			}
			return
		}

		_, err = dst.Write(buf[:n])
		if err != nil {
			p.log.Error(fmt.Sprintf("Write error %s <-> %s. Connection will be closed. Error: %s",src.RemoteAddr().String(), dst.RemoteAddr().String(), err.Error()) )
			return
		}
	}
}

func (p *proxy) protocolError(resp []byte, err error) {
	if resp != nil {
		p.input.Write(resp)
	}
	p.log.Error(err.Error())
}
