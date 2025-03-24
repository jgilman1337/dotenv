package encoder_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/golobby/dotenv/v2"
	//"github.com/stretchr/testify/assert"
)

// Inner struct (linked via pointer)
type FlagBox struct {
	Bool1 bool `env:"BOOL1"`
	Bool2 bool `env:"BOOL2"`
	Bool3 bool `env:"BOOL3"`
	Bool4 bool `env:"BOOL4"`
}

// Outermost struct
type Config struct {
	AppName  string   `env:"APP_NAME"`
	AppPort  int32    `env:"APP_PORT"`
	IPs      []string `env:"IPS"`
	IDs      []int64  `env:"IDS"`
	float    float64  `env:"FLOAT"`
	FlagBox  *FlagBox
	QuoteBox struct {
		Quote1 string `env:"QUOTE1"`
		Quote2 string `env:"QUOTE2"`
		Quote3 string `env:"QUOTE3"`
		Quote4 string `env:"QUOTE4"`
		Quote5 string `env:"QUOTE5"`
	}
}

var (
	//Starter configuration object for the encoder
	cfg = Config{
		AppName: "DotEnv",
		AppPort: 8585,
		IPs:     []string{"192.168.0.1", "192.168.0.2", "192.168.0.3"},
		IDs:     []int64{10, 11, 12, 13, 14},
		float:   3.14,
		FlagBox: &FlagBox{
			Bool1: true,
			Bool2: false,
			Bool3: true,
			Bool4: false,
		},
		QuoteBox: struct {
			Quote1 string `env:"QUOTE1"`
			Quote2 string `env:"QUOTE2"`
			Quote3 string `env:"QUOTE3"`
			Quote4 string `env:"QUOTE4"`
			Quote5 string `env:"QUOTE5"`
		}{
			Quote1: "OK1",
			Quote2: " OK 2 ",
			Quote3: " OK ' 3 ",
			Quote4: " OK \" 4 ",
			Quote5: " OK # 5 ",
		},
	}
)

func TestSaveBytes(t *testing.T) {
	//Setup the encoder and destination
	bytes := bytes.NewBuffer(nil)
	enc := dotenv.NewEncoder(bytes)

	//Encode to the buffer
	if err := enc.Encode(&cfg); err != nil {
		t.Fatal(err)
	}

	//Test for correctness; split the string into an array, delimitating by newlines
	actuals := strings.Split(bytes.String(), "\n")
	for i, actual := range actuals {
		fmt.Printf("L%02d: %s\n", i+1, actual)
	}

	/*
		var c Config
		if err := dotenv.NewDecoder(bytes).Decode(&c); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, cfg, c)
	*/
}
