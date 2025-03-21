// Package dotenv is a lightweight library for loading dot env (.env) files into structs.
package dotenv

import (
	"github.com/golobby/dotenv/v2/pkg/decoder"
	"os"
)

// NewDecoder creates a new instance of decoder.Decoder.
func NewDecoder(file *os.File) *decoder.Decoder {
	return &decoder.Decoder{File: file}
}
