package main

import (
	"bytes"
	"math/rand"
	"testing"
)

func TestGetUser(t *testing.T) {
	pass := []byte(generateRandomString(1,32))
	user := []byte(generateRandomString(1,32))

	input := []byte{0x01, byte(len(user))}
	input = append(input, user...)
	input = append(input, byte(len(pass)))
	input = append(input, pass...)

	result := getUser(input)
	if bytes.Compare(result, user) != 0 {
		t.Errorf("username not expected: %v != %v", user, result)
	}
}

func TestGetPass(t *testing.T) {
	pass := []byte(generateRandomString(1,32))
	user := []byte(generateRandomString(1,32))

	input := []byte{0x01, byte(len(user))}
	input = append(input, user...)
	input = append(input, byte(len(pass)))
	input = append(input, pass...)

	result := getPass(input)
	if bytes.Compare(result, pass) != 0 {
		t.Errorf("password not expected: %v != %v", pass, result )
	}
}

func generateRandomString(min, max int) string{
	return generateString(generateIntInRange(min,max))
}

func generateIntInRange(min, max int) int{
	return rand.Intn(max - min) + min;
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890!@#$%^&*()_+}~"
func generateString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}