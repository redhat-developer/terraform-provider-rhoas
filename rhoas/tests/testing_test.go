package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_RandomString(t *testing.T) {
	length := 20
	rnd := randomString(length)
	assert.Len(t, rnd, length, "got unexpected length")
}

func Test_testAccPreCheck(t *testing.T) {
	testAccPreCheck(t)
}
