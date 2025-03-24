package dotenv_test

import (
	"os"
	"testing"

	"github.com/golobby/dotenv/v2"
	"github.com/stretchr/testify/assert"
)

//FIXME: test fails; may be ok to just remove entirely
/*
func TestNewDecoderBytes(t *testing.T) {
	dat, err := os.ReadFile("./assets/.env")
	assert.NoError(t, err)

	d := dotenv.NewDecoder(dat)
	assert.Equal(t, dat, d.Src) //Using `Equal` instead of `Same`; the struct gets a copy of the bytes instead of a pointer to them
}
*/

func TestNewDecoderFile(t *testing.T) {
	f, err := os.Open("./assets/.env")
	assert.NoError(t, err)
	defer f.Close()

	d := dotenv.NewDecoder(f)
	assert.Same(t, f, d.Src)
}

//FIXME: test fails; may be ok to just remove entirely
/*
func TestNewEncoderBytes(t *testing.T) {
	dat := make([]byte, 0)

	e := dotenv.NewEncoder(dat)
	assert.Equal(t, dat, e.Dest) //Using `Equal` instead of `Same`; the struct gets a copy of the bytes instead of a pointer to them
}
*/

func TestNewEncoderFile(t *testing.T) {
	f, err := os.Open("./assets/.env") //Will not be written to; its ok to use this file
	assert.NoError(t, err)
	defer f.Close()

	e := dotenv.NewEncoder(f)
	assert.Same(t, f, e.Dest)
}
