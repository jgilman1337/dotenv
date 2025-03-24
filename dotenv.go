// Package dotenv is a lightweight library for loading dot env (.env) files into structs.
package dotenv

import (
	"bytes"
	"io"
	"os"

	"github.com/golobby/dotenv/v2/pkg/decoder"
	"github.com/golobby/dotenv/v2/pkg/encoder"
)

// NewDecoder creates a new instance of decoder.Decoder using a byte slice or file descriptor.
func NewDecoder[T ~[]byte | ~*bytes.Buffer | ~*os.File | ~*bytes.Reader](data T) *decoder.Decoder {
	dec := &decoder.Decoder{}
	var src io.Reader

	//Go's generics cannot inference interfaces if 2+ cases fall thru to the same statement; feel free to dedupe the cases for buffer, file, and reader if and when the Go team fixes this
	switch v := any(data).(type) {
	case []byte:
		src = bytes.NewReader(v)
	case *bytes.Buffer:
		src = v
	case *os.File:
		src = v
	case *bytes.Reader:
		src = v
	default:
		panic("unexpected type") //Shouldn't ever be hit
	}

	dec.Src = src
	return dec
}

// NewEncoder creates a new instance of encoder.Encoder using a byte slice or file descriptor.
func NewEncoder[T ~*bytes.Buffer | ~*os.File](data T) *encoder.Encoder {
	enc := &encoder.Encoder{}
	enc.Opts = encoder.DefaultOpts()

	switch v := any(data).(type) {
	case *bytes.Buffer:
		enc.Dest = v
	case *os.File:
		enc.Dest = v
	default:
		panic("unexpected type") //Shouldn't ever be hit
	}

	return enc
}
