package dotenv_test

import (
	"os"
	"testing"

	"github.com/golobby/dotenv/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewDecoderBytes(t *testing.T) {
	dat, err := os.ReadFile("./assets/.env")
	assert.NoError(t, err)

	d := dotenv.NewDecoder(dat)

	assert.Equal(t, dat, d.Bytes) //Using `Equal` instead of `Same`; the struct gets a copy of the bytes instead of a pointer to them
}

func TestNewDecoderFile(t *testing.T) {
	f, err := os.Open("./assets/.env")
	assert.NoError(t, err)

	d := dotenv.NewDecoder(f)

	assert.Same(t, f, d.File)
}
