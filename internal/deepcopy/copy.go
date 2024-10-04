package deepcopy

import (
	"fmt"
	"reflect"
	"unsafe"
)

const (
	flagStickyRO = 1 << 5
	flagEmbedRO  = 1 << 6
	flagRO       = flagStickyRO | flagEmbedRO
)

// Copy performs a deep copy of src and returns the copied value.
func Copy[T any](src T) (T, error) {
	srcValue := reflect.ValueOf(src)
	if srcValue.Kind() == reflect.Invalid {
		var dst T
		return dst, nil
	}

	val, err := copyInternal(srcValue)
	if err != nil {
		var dst T
		return dst, fmt.Errorf("Deep Copy failed, reason: %w", err)
	}

	return val.Interface().(T), nil
}

// MustCopy performs a deep copy of src and returns the copied value
// and panics on any errors encountered.
func MustCopy[T any](src T) T {
	dst, err := Copy(src)
	if err != nil {
		panic(err.Error())
	}
	return dst
}

func copyInternal(src reflect.Value) (reflect.Value, error) {
	switch src.Kind() {
	case reflect.Array:
		return copyArray(src)
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return copyChanOrFuncOrUnsafePointer(src)
	case reflect.Map:
		return copyMap(src)
	case reflect.Pointer:
		return copyPointer(src)
	case reflect.Slice:
		return copySlice(src)
	case reflect.Struct:
		return copyStruct(src)
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String:
		return src, nil
	case reflect.Interface:
		return copyInterface(src)
	default:
		return reflect.Value{}, fmt.Errorf("cannot handle type %s kind %s", src.Type(), src.Kind())
	}
}

func copyArray(src reflect.Value) (reflect.Value, error) {
	if src.Len() == 0 {
		// It is safe to return the source array with zero length arrays
		// as these zero length arrays are immutable.
		return src, nil
	}

	dst := reflect.New(src.Type()).Elem()
	for i := 0; i < src.Len(); i++ {
		dstElement, err := copyInternal(src.Index(i))
		if err != nil {
			return reflect.Value{}, err
		}
		dst.Index(i).Set(dstElement)
	}
	return dst, nil
}

func copyChanOrFuncOrUnsafePointer(src reflect.Value) (reflect.Value, error) {
	if src.IsNil() {
		return src, nil
	}
	return reflect.Value{}, fmt.Errorf("cannot handle non-nil type %s", src.Type())
}

func copyMap(src reflect.Value) (reflect.Value, error) {
	if src.IsNil() {
		return src, nil
	}

	dst := reflect.MakeMap(src.Type())
	for _, key := range src.MapKeys() {
		dstValue, err := copyInternal(src.MapIndex(key))
		if err != nil {
			return reflect.Value{}, err
		}
		dst.SetMapIndex(key, dstValue)
	}
	return dst, nil
}

func copyPointer(src reflect.Value) (reflect.Value, error) {
	if src.IsNil() {
		return src, nil
	}

	dstVal, err := copyInternal(src.Elem())
	if err != nil {
		return reflect.Value{}, err
	}

	dst := reflect.New(src.Type().Elem())
	dst.Elem().Set(dstVal)
	return dst, nil
}

func copySlice(src reflect.Value) (reflect.Value, error) {
	if src.IsNil() {
		return src, nil
	}

	dst := reflect.MakeSlice(src.Type(), src.Len(), src.Cap())
	for i := 0; i < src.Len(); i++ {
		dstElement, err := copyInternal(src.Index(i))
		if err != nil {
			return reflect.Value{}, err
		}
		dst.Index(i).Set(dstElement)
	}
	return dst, nil
}

func copyStruct(src reflect.Value) (reflect.Value, error) {
	dst := reflect.New(src.Type()).Elem()

	for i := 0; i < dst.NumField(); i++ {
		srcF := src.Field(i)
		patchField(&srcF)

		value, err := copyInternal(srcF)
		if err != nil {
			return reflect.Value{}, err
		}

		dstF := dst.Field(i)
		patchField(&dstF)
		dstF.Set(value)
	}
	return dst, nil
}

func copyInterface(src reflect.Value) (reflect.Value, error) {
	if src.IsNil() {
		return src, nil
	}
	return copyInternal(src.Elem())
}

func patchField(src *reflect.Value) {
	// This allows us to access unexported fields for both
	// reading and writing.
	flag := reflect.ValueOf(src).Elem().FieldByName("flag")
	ptrFlag := (*uintptr)(unsafe.Pointer(flag.UnsafeAddr()))
	*ptrFlag = *ptrFlag &^ flagRO
}
