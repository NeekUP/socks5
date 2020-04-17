package main

import (
	"errors"
)

// input positions
const (
	NEG_ARG_VERSION  = 0
	NEG_ARG_NMETHODS = 1
)

type negotiation struct {
	authType AuthType
}

func NewNegotiation(authType AuthType) *negotiation {
	return &negotiation{authType: authType}
}

func (state *negotiation) Receive(input []byte) ([]byte, error) {

	err := state.validate(input)
	if err != nil {
		return nil, err
	}

	version := input[NEG_ARG_VERSION]
	nmethods := int(input[NEG_ARG_NMETHODS])

	for i := 2; i < nmethods+2; i++ {
		if AuthType(input[i]) == state.authType {
			return []byte{version, input[i]}, nil
		}
	}

	return []byte{version, byte(NOT_ACCEPTED)}, nil
}

func (state negotiation) validate(input []byte) error {

	if len(input) < 3 || len(input) > 257 {
		return errors.New("invalid message length")
	}

	if input[NEG_ARG_VERSION] != PROTOCOL_VERSION {
		return errors.New("invalid protocol version")
	}

	if input[NEG_ARG_NMETHODS] == 0 {
		return errors.New("invalid methods number")
	}

	if len(input)-2 != int(input[NEG_ARG_NMETHODS]) {
		return errors.New("invalid methods count")
	}

	return nil
}
