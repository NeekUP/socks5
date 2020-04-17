package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"math/rand"
	"net"
)

// input positions
const (
	CON_ARG_VERSION = 0
	CON_ARG_CMD     = 1
	CON_ARG_RSV     = 2
	CON_ARG_ATYP    = 3
)

// request variables
const (
	CMD_CONNECT = 0x01
	CMD_BIND    = 0x02
	CMD_UDP     = 0x03
	ATYP_IPV4   = 0x01
	ATYP_DOMAIN = 0x03
	ATYP_IPV6   = 0x04
)

// response values
const (
	SUCCESS                    = 0x00
	GENERAL_ERROR              = 0x01
	NOT_ALLOWED_BY_RULSET      = 0x02
	NETWORK_UNREACHABLE        = 0x03
	HOST_UNREACHABLE           = 0x04
	CONNECTION_REFUSED         = 0x05
	TTL_EXPIRED                = 0x06
	COMMAND_NOT_SUPPORTED      = 0x07
	ADDRESS_TYPE_NOT_SUPPORTED = 0x08
)

type connect struct {
	conn   net.Conn
	logger *zap.Logger
	proxy  *proxy
}

func NewRequest(conn net.Conn, proxy *proxy, logger *zap.Logger) *connect {
	return &connect{conn: conn, proxy: proxy, logger: logger}
}

func (state *connect) Receive(input []byte) ([]byte, error) {
	err := state.validate(input)
	if err != nil {
		return nil, err
	}

	addr, err := getAddr(input)
	if err != nil {
		return nil, err
	}

	switch input[CON_ARG_CMD] {
	case CMD_CONNECT:
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			// check cause by error message
			return []byte{PROTOCOL_VERSION, SUCCESS, HOST_UNREACHABLE, 0x01}, err
		}

		if addr, ok := conn.LocalAddr().(*net.TCPAddr); ok {
			state.proxy.output = conn
			return state.response(PROTOCOL_VERSION, SUCCESS, ATYP_IPV4, addr.IP, intToByte(addr.Port)), nil
		}
	case CMD_UDP:
		return state.response(PROTOCOL_VERSION, COMMAND_NOT_SUPPORTED, ATYP_IPV4, []byte{}, []byte{}), nil
	case CMD_BIND:
		return state.response(PROTOCOL_VERSION, COMMAND_NOT_SUPPORTED, ATYP_IPV4, []byte{}, []byte{}), nil
	}

	return state.response(PROTOCOL_VERSION, COMMAND_NOT_SUPPORTED, ATYP_IPV4, []byte{}, []byte{}), nil
}

func intToByte(port int) []byte {
	bndPort := make([]byte, 2)
	binary.BigEndian.PutUint16(bndPort, uint16(port))
	return bndPort
}

func (state *connect) response(protocol, status, atype byte, bndAddr, bndPort []byte) []byte {
	r := []byte{protocol, status, 0x00, atype}
	r = append(r, bndAddr...)
	r = append(r, bndPort...)
	return r
}

func getAddr(input []byte) (string, error) {
	var ip net.IP
	var port int
	isIPv6 := false

	switch input[CON_ARG_ATYP] {
	case ATYP_IPV4:
		if len(input) < 10{
			return "", errors.New("invalid address")
		}
		ip = input[4:8]
		port = int(binary.BigEndian.Uint16(input[8:10]))
	case ATYP_IPV6:
		if len(input) < 22{
			return "", errors.New("invalid address")
		}
		ip = input[4:20]
		port = int(binary.BigEndian.Uint16(input[20:22]))
		isIPv6 = true
	case ATYP_DOMAIN:
		domainLen := int(input[4])
		if len(input) < 4 + domainLen{
			return "", errors.New("invalid domain")
		}
		domain := string(input[5 : 5+domainLen])
		addr, err := net.LookupHost(domain)
		if err != nil {
			return "", err
		}
		if len(addr) == 0 {
			return "", errors.New("domain haven't ip address")
		}
		ip = net.ParseIP(addr[rand.Intn(len(addr))])
		port = int(binary.BigEndian.Uint16(input[5+domainLen : 7+domainLen]))
		isIPv6 = ip.To4() == nil
	}

	if isIPv6 {
		return fmt.Sprintf("[%v]:%v", ip.String(), port), nil
	}
	return fmt.Sprintf("%v:%v", ip.String(), port), nil
}

func (state *connect) validate(input []byte) error {
	if len(input) < 8 {
		return errors.New("invalid message length")
	}

	if input[0] != PROTOCOL_VERSION {
		return errors.New("invalid protocol version")
	}

	if input[1] != CMD_CONNECT && input[1] != CMD_BIND && input[1] != CMD_UDP {
		return errors.New("invalid command")
	}

	if input[3] != ATYP_IPV4 && input[3] != ATYP_IPV6 && input[3] != ATYP_DOMAIN {
		return errors.New("invalid address type")
	}

	if input[3] == ATYP_IPV4 && len(input) != 10{
		return errors.New("invalid message length")
	}

	if input[3] == ATYP_IPV6 && len(input) != 22{
		return errors.New("invalid message length")
	}
	return nil
}
