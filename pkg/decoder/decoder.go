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

// Decode reads a dot env (.env) byte slice or file descriptor and fills the given struct fields.
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

	//Populate the struct
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

	return errors.New("dotenv decode: invalid structure")
} // TODO: might want to make this a utility since it is common to both the encoder and decoder

// feedStruct sets reflected struct fields with the given key/value pairs.
func (d Decoder) feedStruct(s reflect.Value, vars map[string]string) error {
	//Iterate over the fields of the struct
	for i := 0; i < s.NumField(); i++ {
		//Get the current field info
		field := s.Type().Field(i)
		fieldValue := s.Field(i)

		//Check for the `env` struct tag
		if t, exist := field.Tag.Lookup("env"); exist {
			//Case 1: ordinary field; parse the string and populate the corresponding struct field
			if val, exist := vars[t]; exist {
				//Perform the cast to the same type as the target field
				v, err := cast.FromType(val, field.Type)
				if err != nil {
					return fmt.Errorf("dotenv: cannot set `%v` field; err: %v", field.Name, err)
				}

				//Set the value using `unsafe`
				ptr := reflect.NewAt(fieldValue.Type(), unsafe.Pointer(fieldValue.UnsafeAddr())).Elem()
				ptr.Set(reflect.ValueOf(v))
			}
		} else if field.Type.Kind() == reflect.Struct {
			//Case 2: field is an embedded struct; recursively process it
			if err := d.feedStruct(fieldValue, vars); err != nil {
				return err
			}
		} else if field.Type.Kind() == reflect.Ptr {
			//Case 3: field is a pointer to a struct; dereference and recursively process it
			if !fieldValue.IsZero() && fieldValue.Elem().Type().Kind() == reflect.Struct {
				if err := d.feedStruct(fieldValue.Elem(), vars); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
