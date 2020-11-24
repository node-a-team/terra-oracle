package price

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func decodeImpl(data []byte, v interface{}) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, errors.New("obi: decode into non-ptr type")
	}
	ev := rv.Elem()
	switch ev.Kind() {
	case reflect.Uint8:
		val, rem, err := DecodeUnsigned8(data)
		ev.SetUint(uint64(val))
		return rem, err
	case reflect.Uint16:
		val, rem, err := DecodeUnsigned16(data)
		ev.SetUint(uint64(val))
		return rem, err
	case reflect.Uint32:
		val, rem, err := DecodeUnsigned32(data)
		ev.SetUint(uint64(val))
		return rem, err
	case reflect.Uint64:
		val, rem, err := DecodeUnsigned64(data)
		ev.SetUint(uint64(val))
		return rem, err
	case reflect.Int8:
		val, rem, err := DecodeSigned8(data)
		ev.SetInt(int64(val))
		return rem, err
	case reflect.Int16:
		val, rem, err := DecodeSigned16(data)
		ev.SetInt(int64(val))
		return rem, err
	case reflect.Int32:
		val, rem, err := DecodeSigned32(data)
		ev.SetInt(int64(val))
		return rem, err
	case reflect.Int64:
		val, rem, err := DecodeSigned64(data)
		ev.SetInt(int64(val))
		return rem, err
	case reflect.String:
		val, rem, err := DecodeString(data)
		ev.SetString(val)
		return rem, err
	case reflect.Slice:
		if ev.Type().Elem().Kind() == reflect.Uint8 {
			val, rem, err := DecodeBytes(data)
			ev.SetBytes(val)
			return rem, err
		}
		length, rem, err := DecodeUnsigned32(data)
		if err != nil {
			return nil, err
		}
		slice := reflect.MakeSlice(ev.Type(), int(length), int(length))
		for idx := 0; idx < int(length); idx++ {
			var err error
			rem, err = decodeImpl(rem, slice.Index(idx).Addr().Interface())
			if err != nil {
				return nil, err
			}
		}
		ev.Set(slice)
		return rem, nil
	case reflect.Struct:
		rem := data
		for idx := 0; idx < ev.NumField(); idx++ {
			var err error
			rem, err = decodeImpl(rem, ev.Field(idx).Addr().Interface())
			if err != nil {
				return nil, err
			}
		}
		return rem, nil
	default:
		return nil, fmt.Errorf("obi: unsupported value type: %s", ev.Kind())
	}
}

// Decode uses obi encoding scheme to decode the given input(s).
func Decode(data []byte, v ...interface{}) error {
	var err error
	rem := data
	for _, each := range v {
		rem, err = decodeImpl(rem, each)
		if err != nil {
			return err
		}
	}
	if len(rem) != 0 {
		return errors.New("obi: not all data was consumed while decoding")
	}
	return nil
}

// MustDecode uses obi encoding scheme to decode the given input. Panics on error.
func MustDecode(data []byte, v ...interface{}) {
	err := Decode(data, v...)
	if err != nil {
		panic(err)
	}
}

// DecodeUnsigned16 decodes the input bytes into `uint8` and returns the remaining bytes.
func DecodeUnsigned8(data []byte) (uint8, []byte, error) {
	if len(data) < 1 {
		return 0, nil, errors.New("obi: out of range")
	}
	return data[0], data[1:], nil
}

// DecodeUnsigned16 decodes the input bytes into `uint16` and returns the remaining bytes.
func DecodeUnsigned16(data []byte) (uint16, []byte, error) {
	if len(data) < 2 {
		return 0, nil, errors.New("obi: out of range")
	}
	return binary.BigEndian.Uint16(data[:2]), data[2:], nil
}

// DecodeUnsigned32 decodes the input bytes into `uint32` and returns the remaining bytes.
func DecodeUnsigned32(data []byte) (uint32, []byte, error) {
	if len(data) < 4 {
		return 0, nil, errors.New("obi: out of range")
	}
	return binary.BigEndian.Uint32(data[:4]), data[4:], nil
}

// DecodeUnsigned64 decodes the input bytes into `uint64` and returns the remaining bytes.
func DecodeUnsigned64(data []byte) (uint64, []byte, error) {
	if len(data) < 8 {
		return 0, nil, errors.New("obi: out of range")
	}
	return binary.BigEndian.Uint64(data[:8]), data[8:], nil
}

// DecodeSigned8 decodes the input bytes into `uint64` and returns the remaining bytes.
func DecodeSigned8(data []byte) (int8, []byte, error) {
	unsigned, rem, err := DecodeUnsigned8(data)
	return int8(unsigned), rem, err
}

// DecodeSigned16 decodes the input bytes into `uint64` and returns the remaining bytes.
func DecodeSigned16(data []byte) (int16, []byte, error) {
	unsigned, rem, err := DecodeUnsigned16(data)
	return int16(unsigned), rem, err
}

// DecodeSigned32 decodes the input bytes into `uint64` and returns the remaining bytes.
func DecodeSigned32(data []byte) (int32, []byte, error) {
	unsigned, rem, err := DecodeUnsigned32(data)
	return int32(unsigned), rem, err
}

// DecodeSigned64 decodes the input bytes into `uint64` and returns the remaining bytes.
func DecodeSigned64(data []byte) (int64, []byte, error) {
	unsigned, rem, err := DecodeUnsigned64(data)
	return int64(unsigned), rem, err
}

// DecodeBytes decodes the input bytes and returns bytes result and the remaining bytes.
func DecodeBytes(data []byte) ([]byte, []byte, error) {
	length, rem, err := DecodeUnsigned32(data)
	if err != nil {
		return nil, nil, err
	}
	if uint32(len(rem)) < length {
		return nil, nil, errors.New("obi: out of range")
	}
	return rem[:length], rem[length:], nil
}

// DecodeString decodes the input bytes and returns string result and the remaining bytes.
func DecodeString(data []byte) (string, []byte, error) {
	length, rem, err := DecodeUnsigned32(data)
	if err != nil {
		return "", nil, err
	}
	if uint32(len(rem)) < length {
		return "", nil, errors.New("obi: out of range")
	}
	return string(rem[:length]), rem[length:], nil
}

// Encode uses obi encoding scheme to encode the given input into bytes.
func encodeImpl(v interface{}) ([]byte, error) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Uint8:
		return EncodeUnsigned8(uint8(rv.Uint())), nil
	case reflect.Uint16:
		return EncodeUnsigned16(uint16(rv.Uint())), nil
	case reflect.Uint32:
		return EncodeUnsigned32(uint32(rv.Uint())), nil
	case reflect.Uint64:
		return EncodeUnsigned64(uint64(rv.Uint())), nil
	case reflect.Int8:
		return EncodeSigned8(int8(rv.Int())), nil
	case reflect.Int16:
		return EncodeSigned16(int16(rv.Int())), nil
	case reflect.Int32:
		return EncodeSigned32(int32(rv.Int())), nil
	case reflect.Int64:
		return EncodeSigned64(int64(rv.Int())), nil
	case reflect.String:
		return EncodeString(rv.String()), nil
	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return EncodeBytes(rv.Bytes()), nil
		}

		res := EncodeUnsigned32(uint32(rv.Len()))
		for idx := 0; idx < rv.Len(); idx++ {
			each, err := Encode(rv.Index(idx).Interface())
			if err != nil {
				return nil, err
			}
			res = append(res, each...)
		}
		return res, nil
	case reflect.Struct:
		res := []byte{}
		for idx := 0; idx < rv.NumField(); idx++ {
			each, err := Encode(rv.Field(idx).Interface())
			if err != nil {
				return nil, err
			}
			res = append(res, each...)
		}
		return res, nil
	default:
		return nil, fmt.Errorf("obi: unsupported value type: %s", rv.Kind())
	}
}

// Encode uses obi encoding scheme to encode the given input(s) into bytes.
func Encode(v ...interface{}) ([]byte, error) {
	res := []byte{}
	for _, each := range v {
		encoded, err := encodeImpl(each)
		if err != nil {
			return nil, err
		}
		res = append(res, encoded...)
	}
	return res, nil
}

// MustEncode uses obi encoding scheme to encode the given input into bytes. Panics on error.
func MustEncode(v ...interface{}) []byte {
	res, err := Encode(v...)
	if err != nil {
		panic(err)
	}
	return res
}

// EncodeUnsigned8 takes an `uint8` variable and encodes it into a byte array
func EncodeUnsigned8(v uint8) []byte {
	return []byte{v}
}

// EncodeUnsigned16 takes an `uint16` variable and encodes it into a byte array
func EncodeUnsigned16(v uint16) []byte {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, v)
	return bytes
}

// EncodeUnsigned32 takes an `uint32` variable and encodes it into a byte array
func EncodeUnsigned32(v uint32) []byte {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, v)
	return bytes
}

// EncodeUnsigned64 takes an `uint64` variable and encodes it into a byte array
func EncodeUnsigned64(v uint64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, v)
	return bytes
}

// EncodeSigned8 takes an `int8` variable and encodes it into a byte array
func EncodeSigned8(v int8) []byte {
	return EncodeUnsigned8(uint8(v))
}

// EncodeSigned16 takes an `int16` variable and encodes it into a byte array
func EncodeSigned16(v int16) []byte {
	return EncodeUnsigned16(uint16(v))
}

// EncodeSigned32 takes an `int32` variable and encodes it into a byte array
func EncodeSigned32(v int32) []byte {
	return EncodeUnsigned32(uint32(v))
}

// EncodeSigned64 takes an `int64` variable and encodes it into a byte array
func EncodeSigned64(v int64) []byte {
	return EncodeUnsigned64(uint64(v))
}

// EncodeBytes takes a `[]byte` variable and encodes it into a byte array
func EncodeBytes(v []byte) []byte {
	return append(EncodeUnsigned32(uint32(len(v))), v...)
}

// EncodeString takes a `string` variable and encodes it into a byte array
func EncodeString(v string) []byte {
	return append(EncodeUnsigned32(uint32(len(v))), []byte(v)...)
}

func getSchemaImpl(s *strings.Builder, t reflect.Type) error {
	switch t.Kind() {
	case reflect.Uint8:
		s.WriteString("u8")
		return nil
	case reflect.Uint16:
		s.WriteString("u16")
		return nil
	case reflect.Uint32:
		s.WriteString("u32")
		return nil
	case reflect.Uint64:
		s.WriteString("u64")
		return nil
	case reflect.Int8:
		s.WriteString("i8")
		return nil
	case reflect.Int16:
		s.WriteString("i16")
		return nil
	case reflect.Int32:
		s.WriteString("i32")
		return nil
	case reflect.Int64:
		s.WriteString("i64")
		return nil
	case reflect.String:
		s.WriteString("string")
		return nil
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			s.WriteString("bytes")
			return nil
		}
		s.WriteString("[")
		err := getSchemaImpl(s, t.Elem())
		if err != nil {
			return err
		}
		s.WriteString("]")
		return nil
	case reflect.Struct:
		if t.NumField() == 0 {
			return errors.New("obi: empty struct is not supported")
		}
		s.WriteString("{")
		for idx := 0; idx < t.NumField(); idx++ {
			field := t.Field(idx)
			name, ok := field.Tag.Lookup("obi")
			if !ok {
				return fmt.Errorf("obi: no obi tag found for field %s of %s", field.Name, t.Name())
			}
			if idx != 0 {
				s.WriteString(",")
			}
			s.WriteString(name)
			s.WriteString(":")
			err := getSchemaImpl(s, field.Type)
			if err != nil {
				return err
			}
		}
		s.WriteString("}")
		return nil
	default:
		return fmt.Errorf("obi: unsupported value type: %s", t.Kind())
	}
}

// GetSchema returns the compact OBI individual schema of the given value.
func GetSchema(v interface{}) (string, error) {
	s := &strings.Builder{}
	err := getSchemaImpl(s, reflect.TypeOf(v))
	if err != nil {
		return "", err
	}
	return s.String(), nil
}

// MustGetSchema returns the compact OBI individual schema of the given value. Panics on error.
func MustGetSchema(v interface{}) string {
	schema, err := GetSchema(v)
	if err != nil {
		panic(err)
	}
	return schema
}
