package main

import (
	"bytes"
	"errors"
)

type AuthType byte

const (
	NO_AUTH      AuthType = 0x00
	PASS_AUTH    AuthType = 0x02
	NOT_ACCEPTED AuthType = 0xFF
)

const (
	PASS_AUTH_VERSION           = 0x01
	PASS_AUTH_BADREQUEST        = 0x02
	PASS_AUTH_FAIL              = 0x03
	PASS_AUTH_SUCCESS           = 0x00
	MIN_PASSAUTH_MESSAGE_LENGTH = 5
	MAX_PASSAUTH_MESSAGE_LENGTH = 513
)

type passwordAuthentication struct {
	user []byte
	pass []byte
}

func NewPasswordAuthentication(user, pass []byte) *passwordAuthentication {
	return &passwordAuthentication{
		user: user,
		pass: pass,
	}
}

func (state *passwordAuthentication) Receive(input []byte) ([]byte, error) {
	err := state.validate(input)
	if err != nil {
		return []byte{PASS_AUTH_VERSION, PASS_AUTH_BADREQUEST}, err
	}

	user := getUser(input)
	pass := getPass(input)

	if bytes.Compare(user, state.user) != 0 || bytes.Compare(pass, state.pass) != 0 {
		return []byte{PASS_AUTH_VERSION, PASS_AUTH_FAIL}, errors.New("auth fail")
	}

	return []byte{PASS_AUTH_VERSION, PASS_AUTH_SUCCESS}, nil
}

func getUser(input []byte) []byte {
	userlen := int(input[1])
	return input[2 : userlen+2]
}

func getPass(input []byte) []byte {
	userlen := int(input[1])
	return input[3+userlen:]
}

func (state passwordAuthentication) validate(input []byte) error {
	if len(input) < MIN_PASSAUTH_MESSAGE_LENGTH || len(input) > MAX_PASSAUTH_MESSAGE_LENGTH {
		return errors.New("invalid message length")
	}

	if input[0] != PASS_AUTH_VERSION {
		return errors.New("invalid version")
	}

	ulen := int(input[1])
	if ulen == 0 || ulen > len(input)-MIN_PASSAUTH_MESSAGE_LENGTH-1 {
		return errors.New("invalid username length")
	}

	plen := int(input[2+ulen])
	if plen == 0 || plen > len(input)-ulen-3 {
		return errors.New("invalid password length")
	}
	return nil
}
