package main

import (
	"go.uber.org/zap"
	"net"
	"reflect"
	"testing"
)

func Test_connect_Receive(t *testing.T) {
	type fields struct {
		MTU    int
		conn   net.Conn
		logger *zap.Logger
		rconn  net.Conn
		proxy  *proxy
	}

	tests := []struct {
		name    string
		fields  fields
		input   []byte
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &connect{
				conn:   tt.fields.conn,
				logger: tt.fields.logger,
				proxy:  tt.fields.proxy,
			}
			got, err := state.Receive(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Receive() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Receive() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_connect_getAddr(t *testing.T) {

	tests := []struct {
		name    string
		input   []byte
		want    string
		wantErr bool
	}{
		{"valid IPv4",
			[]byte{
				PROTOCOL_VERSION,
				CMD_CONNECT,
				0x00,
				ATYP_IPV4,
				0xAA, 0xAA, 0xAA, 0xAA,
				0x09, 0x10},
			"170.170.170.170:2320",
			false},
		{"valid IPv6",
			[]byte{
				PROTOCOL_VERSION,
				CMD_CONNECT,
				0x00,
				ATYP_IPV6,
				0x4f, 0xfe, 0x29, 0x00, 0x55, 0x45, 0x32, 0x10, 0x20, 0x00, 0xf8, 0xff, 0xfe, 0x21, 0x67, 0xcf,
				0x09, 0x10},
			"[4ffe:2900:5545:3210:2000:f8ff:fe21:67cf]:2320",
			false},
		{"valid domain",
			[]byte{
				PROTOCOL_VERSION,
				CMD_CONNECT,
				0x00,
				ATYP_DOMAIN,
				0x0a, 0x72, 0x65, 0x64, 0x68, 0x61, 0x74, 0x2e, 0x63, 0x6f, 0x6d, // redhat.com
				0x09, 0x10},
			"209.132.183.105:2320",
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getAddr(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAddr() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("getAddr() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_connect_validate(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{"valid IPv4",
			[]byte{
				PROTOCOL_VERSION,
				CMD_CONNECT,
				0x00,
				ATYP_IPV4,
				0xAA, 0xAA, 0xAA, 0xAA,
				0x09, 0x10},
			false},
		{"valid IPv6",
			[]byte{
				PROTOCOL_VERSION,
				CMD_CONNECT,
				0x00,
				ATYP_IPV6,
				0x4f, 0xfe, 0x29, 0x00, 0x55, 0x45, 0x32, 0x10, 0x20, 0x00, 0xf8, 0xff, 0xfe, 0x21, 0x67, 0xcf,
				0x09, 0x10},
			false},
		{"valid domain",
			[]byte{
				PROTOCOL_VERSION,
				CMD_CONNECT,
				0x00,
				ATYP_DOMAIN,
				0x0a, 0x72, 0x65, 0x64, 0x68, 0x61, 0x74, 0x2e, 0x63, 0x6f, 0x6d, // redhat.com
				0x09, 0x10},
			false},
		{"invalid protocol",
			[]byte{
				0x04,
				CMD_CONNECT,
				0x00,
				ATYP_IPV4,
				0xAA, 0xAA, 0xAA, 0xAA,
				0x09, 0x10},
			true},
		{"invalid command",
			[]byte{
				PROTOCOL_VERSION,
				0x04,
				0x00,
				ATYP_IPV4,
				0xAA, 0xAA, 0xAA, 0xAA,
				0x09, 0x10},
			true},
		{"invalid address type",
			[]byte{
				PROTOCOL_VERSION,
				CMD_CONNECT,
				0x00,
				0x10,
				0xAA, 0xAA, 0xAA, 0xAA,
				0x09, 0x10},
			true},
		{"invalid IPv4 address",
			[]byte{
				PROTOCOL_VERSION,
				CMD_CONNECT,
				0x00,
				0x10,
				0xAA, 0xAA, 0xAA,
				0x09, 0x10},
			true},
		{"invalid IPv4 port address",
			[]byte{
				PROTOCOL_VERSION,
				CMD_CONNECT,
				0x00,
				0x10,
				0xAA, 0xAA, 0xAA, 0xAA,
				0x09},
			true},
		{"invalid IPv6 address",
			[]byte{
				PROTOCOL_VERSION,
				CMD_CONNECT,
				0x00,
				ATYP_IPV6,
				0x4f, 0xfe, 0x29, 0x00, 0x55, 0x45, 0x32, 0x10, 0x20, 0x00, 0xf8, 0xff, 0xfe, 0x21, 0x67,
				0x09, 0x10},
			true},
		{"invalid IPv6 address port",
			[]byte{
				PROTOCOL_VERSION,
				CMD_CONNECT,
				0x00,
				ATYP_IPV6,
				0x4f, 0xfe, 0x29, 0x00, 0x55, 0x45, 0x32, 0x10, 0x20, 0x00, 0xf8, 0xff, 0xfe, 0x21, 0x67,
				0x09},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &connect{}
			if err := state.validate(tt.input); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
