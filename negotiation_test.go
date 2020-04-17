package main

import (
	"reflect"
	"testing"
)

func Test_negotiation_Receive(t *testing.T) {
	tests := []struct {
		name    string
		auth    AuthType
		input   []byte
		want    []byte
		wantErr bool
	}{
		{"pass auth", PASS_AUTH, []byte{PROTOCOL_VERSION, 0x01, byte(PASS_AUTH)}, []byte{PROTOCOL_VERSION, byte(PASS_AUTH)}, false},
		{"no auth", NO_AUTH, []byte{PROTOCOL_VERSION, 0x01, byte(NO_AUTH)}, []byte{PROTOCOL_VERSION, byte(NO_AUTH)}, false},
		{"not accepted auth NO_AUTH", NO_AUTH, []byte{PROTOCOL_VERSION, 0x01, byte(PASS_AUTH)}, []byte{PROTOCOL_VERSION, byte(NOT_ACCEPTED)}, false},
		{"not accepted auth PASS_AUTH", PASS_AUTH, []byte{PROTOCOL_VERSION, 0x01, byte(NO_AUTH)}, []byte{PROTOCOL_VERSION, byte(NOT_ACCEPTED)}, false},
		{"multiple variants PASS_AUTH", PASS_AUTH, []byte{PROTOCOL_VERSION, 0x02, byte(NO_AUTH),byte(PASS_AUTH)}, []byte{PROTOCOL_VERSION, byte(PASS_AUTH)}, false},
		{"multiple variants NO_AUTH", NO_AUTH, []byte{PROTOCOL_VERSION, 0x02, byte(NO_AUTH),byte(PASS_AUTH)}, []byte{PROTOCOL_VERSION, byte(NO_AUTH)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &negotiation{
				authType: tt.auth,
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

func Test_negotiation_validate(t *testing.T) {
	tests := []struct {
		name    string
		auth    AuthType
		input   []byte
		wantErr bool
	}{
		{"valid pass auth", PASS_AUTH, []byte{PROTOCOL_VERSION, 0x01, byte(PASS_AUTH)}, false},
		{"valid no auth", PASS_AUTH, []byte{PROTOCOL_VERSION, 0x01, byte(NO_AUTH)}, false},
		{"valid multiple auth", PASS_AUTH, []byte{PROTOCOL_VERSION, 0x02, byte(PASS_AUTH), byte(NO_AUTH)}, false},
		{"without auth type", PASS_AUTH, []byte{PROTOCOL_VERSION, 0x01}, true},
		{"only version", PASS_AUTH, []byte{PROTOCOL_VERSION}, true},
		{"empty", PASS_AUTH, []byte{}, true},
		{"invalid version", PASS_AUTH, []byte{0x04, 0x01, byte(PASS_AUTH)}, true},
		{"invalid methods number", PASS_AUTH, []byte{PROTOCOL_VERSION, 0x00, byte(PASS_AUTH)}, true},
		{"invalid methods count", PASS_AUTH, []byte{PROTOCOL_VERSION, 0x01, byte(PASS_AUTH), byte(NO_AUTH)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := negotiation{
				authType: tt.auth,
			}
			if err := state.validate(tt.input); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
