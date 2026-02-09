package pod

import (
	"fmt"
	"io"
	"kaiju/klib"
	"math"
	"reflect"
	"slices"
	"strings"
)

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) Encoder {
	return Encoder{w}
}

func (e Encoder) Encode(from any) error {
	// There exists a global `registry` which is a key (string) value (any) pair.
	// This registry allows us to map what concrete type needs to be dynamically
	// created when we decode the buffer.
	//
	// First we need to collect all the keys being used by the `from` type
	typeLookup, err := extractUsedRegistryKeys(from)
	if err != nil {
		return err
	}
	// Next, we will encode all the keys into the header. This will make a look
	// up table at the beginning of the file. This is to reduce the amount of
	// data being saved into the binary blob. Instead of storing the key every
	// time it's used, it can reference this table by id.
	if err = klib.BinaryWriteStringSlice(e.w, typeLookup); err != nil {
		return err
	}
	// Likewise, we will collect all the unique field names within the pod
	// structure. This will make it so that fields can move around or change
	// and it won't break the encoding/decoding.
	fieldLookup := extractUsedFieldKeys(from)
	if err = klib.BinaryWriteStringSlice(e.w, fieldLookup); err != nil {
		return err
	}
	// Next, we begin to serialize the struct's value, this will recursively
	// encode all the fields in the struct.
	return e.encodeValue(reflect.ValueOf(from), typeLookup, fieldLookup)
}

func (e Encoder) encodeValue(val reflect.Value, typeLookup, fieldLookup []string) error {
	// We don't encode empty arrays, slices, or maps
	switch val.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		if val.Len() == 0 {
			return nil
		}
	}
	// First, we need to determine what we are about to encode. We use an int
	// for this identification. Negative numbers are reserved for Go primitives
	// while 0+ are directly mapped to the typeLookup. We write this type id so
	// that the decoder knows what it should be generating
	if err := e.encodeTypeId(val, typeLookup); err != nil {
		return err
	}
	// Next, we need to encode the name of the field
	if err := e.encodeFields(val, typeLookup, fieldLookup); err != nil {
		return err
	}
	return nil
}

func (e Encoder) encodeTypeId(val reflect.Value, typeLookup []string) error {
	t := val.Type()
	kindType := uint8(0)
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		kindType = kindTypeSliceArray
	default:
		qn := qualifiedName(t)
		k := slices.Index(typeLookup, qn)
		if k < 0 || k > math.MaxUint8 {
			return fmt.Errorf("encoding type '%s' was never registered with pod", qn)
		}
		kindType = uint8(k)
	}
	return klib.BinaryWrite(e.w, kindType)
}

func (e Encoder) encodeFields(val reflect.Value, typeLookup, fieldLookup []string) error {
	fieldCount := uint8(0)
	// Detect a generated structure
	k := val.Kind()
	if k > reflect.UnsafePointer {
		val = reflect.Indirect(val)
	}
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		count := val.Len()
		if err := klib.BinaryWriteInt(e.w, count); err != nil {
			return err
		}
		for i := range count {
			v := val.Index(i)
			for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
				v = v.Elem()
			}
			if err := e.encodeValue(v, typeLookup, fieldLookup); err != nil {
				return err
			}
		}
	case reflect.Struct:
		const maxFields = math.MaxUint8
		t := val.Type()
		count := t.NumField()
		if count > maxFields {
			return fmt.Errorf("pod encoding only supports up to %d fields per-struct, '%s' has %d", maxFields, qualifiedName(t), count)
		}
		fieldCount = uint8(count)
		for i := range int(fieldCount) {
			f := val.Field(i)
			switch f.Kind() {
			case reflect.Pointer, reflect.Interface, reflect.Chan,
				reflect.Func, reflect.UnsafePointer:
				fieldCount--
			case reflect.Array, reflect.Slice, reflect.Map:
				// We don't encode empty arrays, slices, or maps
				if f.Len() == 0 {
					fieldCount--
				}
			}
		}
		if err := klib.BinaryWrite(e.w, fieldCount); err != nil {
			return err
		}
		for i := range count {
			f := val.Field(i)
			switch f.Kind() {
			case reflect.Pointer, reflect.Interface, reflect.Chan,
				reflect.Func, reflect.UnsafePointer:
				continue
			case reflect.Array, reflect.Slice, reflect.Map:
				// We don't encode empty arrays, slices, or maps
				if f.Len() == 0 {
					continue
				}
			}
			// First, encode the field lookup id
			fidx := slices.Index(fieldLookup, t.Field(i).Name)
			if fidx < 0 {
				return fmt.Errorf("field '%s' not found in field lookup", t.Field(i).Name)
			}
			if err := klib.BinaryWrite(e.w, uint16(fidx)); err != nil {
				return err
			}
			// Then encode the field value
			if err := e.encodeValue(f, typeLookup, fieldLookup); err != nil {
				return err
			}
		}
	default:
		switch val.Kind() {
		case reflect.Int:
			val = reflect.ValueOf(int32(val.Interface().(int)))
		case reflect.String:
			return klib.BinaryWriteString(e.w, val.String())
		}
		return klib.BinaryWrite(e.w, val.Interface())
	}
	return nil
}

// extractUsedRegistryKeys will recursively go through all fields of the `from`
// argument and uniquely collect all keys that have been registered to pod
func extractUsedRegistryKeys(from any) ([]string, error) {
	unique := make(map[string]struct{})
	collectQualifiedNames(reflect.ValueOf(from), unique)
	structKeyMap := make([]string, 0, len(unique))
	for k := range unique {
		if _, ok := registry.Load(k); !ok {
			return structKeyMap, fmt.Errorf("expected '%s' to have been registered for pod encoding", k)
		}
		structKeyMap = append(structKeyMap, k)
		if len(structKeyMap) == int(kindTypeSliceArray) {
			return structKeyMap, fmt.Errorf("too many types for pod encoding on '%s', max allowed unique types are %d", k, kindTypeSliceArray-1)
		}
	}
	return structKeyMap, nil
}

func extractUsedFieldKeys(from any) []string {
	unique := make(map[string]struct{})
	collectQualifiedFieldNames(reflect.ValueOf(from), unique)
	return klib.MapKeys(unique)
}

// collectQualifiedNames recursively walks through struct fields and records
// the qualified name of each exported, non‑pointer, non‑interface field type.
func collectQualifiedNames(src reflect.Value, set map[string]struct{}) {
	switch src.Kind() {
	case reflect.Interface, reflect.Pointer:
		// If the interface value is nil, there is nothing to collect.
		if src.IsNil() {
			return
		}
		// Get the concrete value stored in the interface and recurse.
		collectQualifiedNames(src.Elem(), set)
		return
	case reflect.Slice, reflect.Array:
		if src.Len() == 0 {
			// We don't pack empty arrays anyway
			return
		}
		for i := range src.Len() {
			collectQualifiedNames(src.Index(i), set)
		}
		return
	case reflect.Struct:
		t := src.Type()
		qn := qualifiedName(t)
		set[qn] = struct{}{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			// Unexported fields are ignored
			if f.PkgPath != "" {
				continue
			}
			collectQualifiedNames(src.Field(i), set)
		}
	default:
		qn := qualifiedName(src.Type())
		set[qn] = struct{}{}
	}
}

func collectQualifiedFieldNames(src reflect.Value, set map[string]struct{}) {
	switch src.Kind() {
	case reflect.Interface, reflect.Pointer:
		// If the interface value is nil, there is nothing to collect.
		if src.IsNil() {
			return
		}
		// Get the concrete value stored in the interface and recurse.
		collectQualifiedFieldNames(src.Elem(), set)
		return
	case reflect.Slice, reflect.Array:
		if src.Len() == 0 {
			// We don't pack empty arrays anyway
			return
		}
		for i := range src.Len() {
			collectQualifiedFieldNames(src.Index(i), set)
		}
		return
	case reflect.Struct:
		t := src.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			set[f.Name] = struct{}{}
			// Unexported fields are ignored
			if f.PkgPath != "" {
				continue
			}
			collectQualifiedFieldNames(src.Field(i), set)
		}
	}
}

// qualifiedName returns a human‑readable name for the given type.
//
// It builds the name using the package path of the type, if available.
// If the type's package path is empty (e.g., for built‑in types or types
// defined in the main package), the function simply returns the type's name.
// Otherwise, it extracts the last element of the package path (the package
// name) and concatenates it with the type's name, separated by a dot.
//
// Parameters:
//
//	t - the reflect.Type whose qualified name is to be generated.
//
// Returns:
//
//	A string in the form "PackageName.TypeName" when the package path is
//	present, or just "TypeName" when the package path is empty.
func qualifiedName(t reflect.Type) string {
	if t.Kind() == reflect.Interface {
		return ""
	} else if path := t.PkgPath(); path == "" {
		name := t.Name()
		if name == "" {
			registry.Range(func(k, v any) bool {
				if v == t {
					name = k.(string)
					return false
				}
				return true
			})
		}
		return name
	} else {
		p := strings.Split(path, "/")
		return p[len(p)-1] + "." + t.Name()
	}
}
