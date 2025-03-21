// Package dotenv is a lightweight library for loading dot env (.env) files into structs.
package dotenv

import (
	"os"

	"github.com/golobby/dotenv/v2/pkg/decoder"
)

// NewDecoder creates a new instance of decoder.Decoder using a byte slice or file.
func NewDecoder[T []byte | *os.File](data T) *decoder.Decoder {
	switch v := any(data).(type) {
	case []byte:
		return &decoder.Decoder{Bytes: v}
	case *os.File:
		return &decoder.Decoder{File: v}
	default:
		return nil //Shouldn't ever be hit
	}
}
