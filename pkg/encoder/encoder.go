package encoder

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"unsafe"

	"github.com/spf13/cast"
)

type Encoder struct {
	Dest io.Writer

	Opts EncoderOpts
}

// Encode reads a given struct, converts it to a map, and writes it to a dot env (.env) to a byte slice and/or file.
func (e Encoder) Encode(structure interface{}) error {
	//Ensure the encoder has a data source to write to
	if e.Dest == nil {
		return fmt.Errorf("no valid data sources could be found for the encoder")
	}

	//Write the struct data to a map of strings
	items, err := e.feed(structure)
	if err != nil {
		return err
	}

	//Write the map to the output objects set by the encoder
	for i, item := range items {
		//Create the initial KV line
		kvSep := ""
		if e.Opts.SpaceAroundKV {
			kvSep = " "
		}
		entry := item.Key + kvSep + "=" + kvSep + item.Value
		lineDelim := "\n" //It is assumed that LF is ok on the host OS

		//Create the metadata line
		meta := ""
		mpath := ""
		mtype := ""
		if e.Opts.IncludePath || e.Opts.IncludeTyping {
			//Override the "BlankLinesBetweenKV" value to make it always true
			e.Opts.BlankLinesBetweenKV = true

			//Init the path and type sections if the user opted to include them
			if e.Opts.IncludePath {
				mpath = "Path: " + item.Path
			}
			if e.Opts.IncludeTyping {
				mtype = "Type: " + item.Datatype
			}

			//If both were requested, add a delimiter
			metaPrefix := "# "
			mdelim := ""
			if e.Opts.IncludePath && e.Opts.IncludeTyping {
				mdelim = "\n" + metaPrefix
				if e.Opts.MinifyPTInfo {
					mdelim = "; "
				}
			}
			meta = metaPrefix + mpath + mdelim + mtype + lineDelim
		}

		//Add a line terminator beforehand if this line succeeds a previous one
		line := meta + entry
		if i > 0 {
			delim := lineDelim
			if e.Opts.BlankLinesBetweenKV {
				delim += delim
			}
			line = delim + line
		}

		//Write the line
		_, err := e.Dest.Write([]byte(line))
		if err != nil {
			return err
		}
	}

	return nil
}

// feed sets key/value pairs with the given struct fields.
func (e Encoder) feed(structure interface{}) ([]_EnvLine, error) {
	inputType := reflect.TypeOf(structure)
	if inputType != nil {
		if inputType.Kind() == reflect.Ptr {
			if inputType.Elem().Kind() == reflect.Struct {
				//Get the number of fields in the struct; allows for efficient array allocs
				//Only accounts for the topmost struct; will not pre-alloc for nested structs
				slen := inputType.Elem().NumField()
				items := make([]_EnvLine, 0, slen)

				//Get the initial path and begin processing
				//The struct is always a pointer, so strip out the first char of the path
				pname := inputType.String()[strings.LastIndex(inputType.String(), ".")+1:] + "."
				return items, e.feedMap(reflect.ValueOf(structure).Elem(), pname, &items)
			}
		}
	}

	return nil, errors.New("dotenv encode: invalid structure")
}

// feedMap sets key/value pairs with the given reflected struct fields.
func (e Encoder) feedMap(s reflect.Value, path string, items *[]_EnvLine) error {
	//Iterate over the fields of the struct
	for i := 0; i < s.NumField(); i++ {
		//Get the current field info
		field := s.Type().Field(i)
		fieldValue := s.Field(i)

		//Check for the `env` struct tag
		if t, exist := field.Tag.Lookup("env"); exist {
			//Case 1: ordinary field; convert it to a string and save it to the map
			strval, err := e.cast2String(fieldValue)
			if err != nil {
				return fmt.Errorf("cannot convert field `%v` to string: %v", field.Name, err)
			}
			dt := fieldValue.Type().String()

			*items = append(*items, _EnvLine{t, strval, dt, path + field.Name})
		} else if field.Type.Kind() == reflect.Struct || field.Type.Kind() == reflect.Ptr {
			//Case 2/3: field is an embedded struct; recursively process it
			/*
				Embedded structs don't get an `env` struct tag since dotenv files are flat,
				and have no concept of nesting, unlike TOML. Because of this, it is assumed
				that this section will only process structs and pointers to structs.
			*/

			//Dereference the struct if its a nonzero struct pointer
			estruct := fieldValue
			if field.Type.Kind() == reflect.Ptr &&
				!fieldValue.IsZero() && fieldValue.Elem().Type().Kind() == reflect.Struct {
				estruct = fieldValue.Elem()
			} //TODO: what about non-struct pointers? What about if the user neglected to label an an ordinary field with the `env` struct tag?

			//Recursively process the struct
			newp := path + field.Name + "."
			if err := e.feedMap(estruct, newp, items); err != nil {
				return err
			}
		}

		//TODO: remove this old impl of cases 2/3 once its confirmed the above is stable
		/*
			} else if field.Type.Kind() == reflect.Struct || field.Type.Kind() == reflect.Ptr {
						//Case 2/3: field is an embedded struct; recursively process it
						if err := e.feedMap(s.Field(i), s.Type().Name()+".", items); err != nil {
							return err
						}
					} else if field.Type.Kind() == reflect.Ptr {
						//Case 3: field is a pointer to a struct; dereference and recursively process it
						if !fieldValue.IsZero() && fieldValue.Elem().Type().Kind() == reflect.Struct {
							if err := e.feedMap(fieldValue.Elem(), path+s.Type().Name()+".", items); err != nil {
								return err
							}
						}
					}
		*/
	}

	return nil
}

// Utility to cast a reflected type into a string, including slices; uses `spf13/cast` internally.
func (e Encoder) cast2String(v reflect.Value) (string, error) {
	//Check for arrays and slices
	kind := v.Kind()
	if kind == reflect.Slice || kind == reflect.Array {
		//Process each item recursively
		strs := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			var err error
			strs[i], err = e.cast2String(v.Index(i))
			if err != nil {
				return "", err
			}
		}

		//Emit the built array as a comma delimited string
		sep := ","
		if e.Opts.SpacesInArrs {
			sep += " "
		}
		return strings.Join(strs, sep), nil
	}

	//Cast the item to a string
	str, err := cast.ToStringE(getRealValue(v))
	if err != nil {
		return "", err
	}

	//Check if the string is prefixed and/or suffixed with spaces
	//If so, quote the string and escape all double quotes
	if strings.HasPrefix(str, " ") || strings.HasSuffix(str, " ") {
		str = strings.ReplaceAll(str, "\"", "\\\"")
		str = "\"" + str + "\""
	}

	return str, nil
}

// Gets the value of a reflected field via `unsafe`. This allows processing of unexported fields.
func getRealValue(v reflect.Value) any {
	ptr := reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
	return ptr.Interface()
}
