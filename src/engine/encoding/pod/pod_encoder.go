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
	qn := qualifiedName(t)
	kindType := uint8(0)
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		kindType = kindTypeSliceArray
	default:
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
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		count := val.Len()
		if err := klib.BinaryWriteInt(e.w, count); err != nil {
			return err
		}
		for i := range count {
			if err := e.encodeValue(val.Index(i), typeLookup, fieldLookup); err != nil {
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
			switch val.Field(i).Kind() {
			case reflect.Pointer, reflect.Interface, reflect.Chan,
				reflect.Func, reflect.UnsafePointer:
				fieldCount--
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
			}
			// First, encode the field lookup id
			idx := uint16(slices.Index(fieldLookup, t.Field(i).Name))
			if err := klib.BinaryWrite(e.w, idx); err != nil {
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
			return klib.BinaryWriteString(e.w, val.Interface().(string))
		}
		return klib.BinaryWrite(e.w, val.Interface())
	}
	return nil
}

// extractUsedRegistryKeys will recursively go through all fields of the `from`
// argument and uniquely collect all keys that have been registered to pod
func extractUsedRegistryKeys(from any) ([]string, error) {
	unique := make(map[string]struct{})
	collectQualifiedNames(reflect.TypeOf(from), unique)
	structKeyMap := make([]string, 0, len(unique))
	for k := range unique {
		if _, ok := registry.Load(k); !ok {
			return structKeyMap, fmt.Errorf("expected '%s' to have been registered for kob encoding", k)
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
	collectQualifiedFieldNames(reflect.TypeOf(from), unique)
	return klib.MapKeys(unique)
}

// collectQualifiedNames recursively walks through struct fields and records
// the qualified name of each exported, non‑pointer, non‑interface field type.
func collectQualifiedNames(t reflect.Type, set map[string]struct{}) {
	if name := qualifiedName(t); name != "" {
		set[name] = struct{}{}
	}
	kind := pullInnerKind(t)
	if kind != reflect.Struct {
		set[kind.String()] = struct{}{}
		return
	}
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		for t.Kind() != reflect.Struct {
			t = t.Elem()
		}
		collectQualifiedNames(t, set)
		return
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		ft := f.Type
		if ft.Kind() == reflect.Ptr || ft.Kind() == reflect.Interface {
			continue
		}
		collectQualifiedNames(ft, set)
	}
}

func collectQualifiedFieldNames(t reflect.Type, set map[string]struct{}) {
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		t = t.Elem()
		collectQualifiedFieldNames(t, set)
		return
	}
	if t.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		ft := f.Type
		if ft.Kind() == reflect.Ptr || ft.Kind() == reflect.Interface {
			continue
		}
		set[f.Name] = struct{}{}
		collectQualifiedFieldNames(ft, set)
	}
}

// pullInnerKind returns the underlying element kind for container types.
// For reflect.Array, reflect.Slice, and reflect.Map, it returns the kind of the
// element type (t.Elem().Kind()). For all other types, it returns the type's
// own kind (t.Kind()). This helper is useful when encoding or decoding
// values that may be wrapped in collection types, allowing callers to work
// with the primitive kind of the stored elements.
func pullInnerKind(t reflect.Type) reflect.Kind {
	switch t.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		return pullInnerKind(t.Elem())
	default:
		return t.Kind()
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
	if path := t.PkgPath(); path == "" {
		return t.Name()
	} else {
		p := strings.Split(path, "/")
		return p[len(p)-1] + "." + t.Name()
	}
}
