package decoder

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"unsafe"

	"github.com/golobby/cast"
)

type Decoder struct {
	Bytes []byte
	File  *os.File
}

// Decode reads a dot env (.env) byte slice and fills the given struct fields.
func (d Decoder) Decode(structure interface{}) error {
	var datSrc io.Reader

	//Try to read the byte slice field first, otherwise read the file field instead
	if d.Bytes != nil {
		datSrc = bytes.NewBuffer(d.Bytes)
	} else if d.File != nil {
		datSrc = d.File
	} else {
		return fmt.Errorf("no valid data sources could be found for the decoder")
	}

	//Read in the dotenv data source
	kvs, err := d.read(datSrc)
	if err != nil {
		return err
	}

	if err := d.feed(structure, kvs); err != nil {
		return err
	}

	return nil
}

// read scans a dot env (.env) data source and extracts its key/value pairs.
func (d Decoder) read(dat io.Reader) (map[string]string, error) {
	kvs := map[string]string{}
	scanner := bufio.NewScanner(dat)

	for i := 1; scanner.Scan(); i++ {
		if k, v, err := d.parse(scanner.Text()); err != nil {
			return nil, fmt.Errorf("dotenv: error in line %v; err: %v", i, err)
		} else if k != "" {
			kvs[k] = v
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("dotenv: error when scanning file; err: %v", err)
	}

	return kvs, nil
}

// parse extracts a key/value pair from the given dot env (.env) single line.
func (d Decoder) parse(line string) (string, string, error) {
	ln := strings.TrimSpace(line)
	kv := []string{"", ""}
	pi := 0
	iq := false
	qt := "'"

	for i := 0; i < len(ln); i++ {
		if string(ln[i]) == "#" && pi == 0 {
			break
		}

		if string(ln[i]) == "#" && pi == 1 && !iq {
			break
		}

		if string(ln[i]) == "=" && pi == 0 {
			pi = 1
			continue
		}

		if string(ln[i]) == " " && pi == 1 {
			if !iq && kv[pi] == "" {
				continue
			}
		}

		if (string(ln[i]) == "\"" || string(ln[i]) == "'") && pi == 1 {
			if kv[pi] == "" {
				iq = true
				qt = string(ln[i])
				continue
			} else if iq && qt == string(ln[i]) {
				break
			}
		}

		kv[pi] += string(ln[i])
	}

	kv[0] = strings.TrimSpace(kv[0])
	if !iq {
		kv[1] = strings.TrimSpace(kv[1])
	}

	if (pi == 0 && kv[0] != "") || (pi == 1 && kv[0] == "") {
		return "", "", fmt.Errorf("dotenv: invalid syntax")
	}

	return kv[0], kv[1], nil
}

// feed sets struct fields with the given key/value pairs.
func (d Decoder) feed(structure interface{}, kvs map[string]string) error {
	inputType := reflect.TypeOf(structure)
	if inputType != nil {
		if inputType.Kind() == reflect.Ptr {
			if inputType.Elem().Kind() == reflect.Struct {
				return d.feedStruct(reflect.ValueOf(structure).Elem(), kvs)
			}
		}
	}

	return errors.New("dotenv: invalid structure")
}

// feedStruct sets reflected struct fields with the given key/value pairs.
func (d Decoder) feedStruct(s reflect.Value, vars map[string]string) error {
	for i := 0; i < s.NumField(); i++ {
		if t, exist := s.Type().Field(i).Tag.Lookup("env"); exist {
			if val, exist := vars[t]; exist {
				v, err := cast.FromType(val, s.Type().Field(i).Type)
				if err != nil {
					return fmt.Errorf("dotenv: cannot set `%v` field; err: %v", s.Type().Field(i).Name, err)
				}

				ptr := reflect.NewAt(s.Field(i).Type(), unsafe.Pointer(s.Field(i).UnsafeAddr())).Elem()
				ptr.Set(reflect.ValueOf(v))
			}
		} else if s.Type().Field(i).Type.Kind() == reflect.Struct {
			if err := d.feedStruct(s.Field(i), vars); err != nil {
				return err
			}
		} else if s.Type().Field(i).Type.Kind() == reflect.Ptr {
			//if s.Field(i).IsZero() == false && s.Field(i).Elem().Type().Kind() == reflect.Struct {
			if !s.Field(i).IsZero() && s.Field(i).Elem().Type().Kind() == reflect.Struct {
				if err := d.feedStruct(s.Field(i).Elem(), vars); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
