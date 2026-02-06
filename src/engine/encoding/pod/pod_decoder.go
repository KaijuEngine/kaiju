package pod

import (
	"errors"
	"fmt"
	"io"
	"kaiju/klib"
	"reflect"
)

type Decoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) Decoder {
	return Decoder{r}
}

func (d Decoder) Decode(into any) error {
	// Read the type lookup table from the header
	typeLookup, err := klib.BinaryReadStringSlice(d.r)
	if err != nil {
		return fmt.Errorf("failed to read type lookup table: %w", err)
	}
	// Read the field lookup table from the header
	fieldLookup, err := klib.BinaryReadStringSlice(d.r)
	if err != nil {
		return fmt.Errorf("failed to read field lookup table: %w", err)
	}
	// Decode the value into the target
	val := reflect.ValueOf(into)
	if val.Kind() != reflect.Ptr {
		return errors.New("into must be a pointer")
	}
	val = val.Elem()
	return d.decodeValue(val, typeLookup, fieldLookup)
}

func (d Decoder) decodeValue(val reflect.Value, typeLookup, fieldLookup []string) error {
	// Read the type id to determine what type we're decoding
	var typeId uint8
	if err := klib.BinaryRead(d.r, &typeId); err != nil {
		return fmt.Errorf("failed to read type id: %w", err)
	}
	// Decode based on the type id
	if typeId == kindTypeSliceArray {
		return d.decodeSliceOrArray(val, typeLookup, fieldLookup)
	}
	if val.Kind() == reflect.Interface {
		key := typeLookup[typeId]
		if r, ok := registry.Load(key); !ok {
			return fmt.Errorf("missing registration in POD for '%s'", key)
		} else {
			ival := reflect.New(r.(reflect.Type)).Elem()
			err := d.decodeFieldsForType(ival, typeLookup, fieldLookup)
			if err != nil {
				return err
			}
			val.Set(ival)
			return nil
		}
	} else {
		return d.decodeFieldsForType(val, typeLookup, fieldLookup)
	}
}

func (d Decoder) decodeSliceOrArray(val reflect.Value, typeLookup, fieldLookup []string) error {
	// Read the count of elements
	count, err := klib.BinaryReadInt(d.r)
	if err != nil {
		return fmt.Errorf("failed to read slice/array count: %w", err)
	}
	// For slices, we need to allocate the slice
	if val.Kind() == reflect.Slice {
		val.Set(reflect.MakeSlice(val.Type(), int(count), int(count)))
	} else if val.Kind() != reflect.Array {
		return fmt.Errorf("expected slice or array, got %v", val.Kind())
	}
	// Decode each element
	for i := 0; i < int(count); i++ {
		elemVal := val.Index(i)
		if err := d.decodeValue(elemVal, typeLookup, fieldLookup); err != nil {
			return fmt.Errorf("failed to decode element %d: %w", i, err)
		}
	}
	return nil
}

func (d Decoder) decodeFieldsForType(val reflect.Value, typeLookup, fieldLookup []string) error {
	switch val.Kind() {
	case reflect.Struct:
		return d.decodeStruct(val, typeLookup, fieldLookup)
	case reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String:
		return d.decodePrimitive(val)
	default:
		return fmt.Errorf("unsupported kind for decoding: %v", val.Kind())
	}
}

func (d Decoder) decodeStruct(val reflect.Value, typeLookup, fieldLookup []string) error {
	// Read the field count
	var fieldCount uint8
	if err := klib.BinaryRead(d.r, &fieldCount); err != nil {
		return fmt.Errorf("failed to read field count: %w", err)
	}
	// Read each field
	for i := 0; i < int(fieldCount); i++ {
		// Read the field index from the field lookup table
		var fieldIdx uint16
		if err := klib.BinaryRead(d.r, &fieldIdx); err != nil {
			return fmt.Errorf("failed to read field index %d: %w", i, err)
		}
		if int(fieldIdx) >= len(fieldLookup) {
			return fmt.Errorf("field index %d out of range (lookup table size: %d)", fieldIdx, len(fieldLookup))
		}
		fieldName := fieldLookup[fieldIdx]
		// Find the field in the struct by name
		f, ok := val.Type().FieldByName(fieldName)
		if !ok {
			return fmt.Errorf("field '%s' not found in struct %v", fieldName, val.Type())
		}
		fieldVal := val.FieldByIndex(f.Index)
		// Decode the field value
		if err := d.decodeValue(fieldVal, typeLookup, fieldLookup); err != nil {
			return fmt.Errorf("failed to decode field '%s': %w", fieldName, err)
		}
	}
	return nil
}

func (d Decoder) decodePrimitive(val reflect.Value) error {
	if val.Kind() == reflect.String {
		str, err := klib.BinaryReadString(d.r)
		if err != nil {
			return err
		}
		val.SetString(str)
	} else {
		ptr := reflect.New(val.Type())
		if err := klib.BinaryRead(d.r, ptr.Interface()); err != nil {
			return fmt.Errorf("failed to read primitive value: %w", err)
		}
		val.Set(ptr.Elem())
	}
	return nil
}
