package deepcopy

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/tuxdudehomelab/homelab/internal/newutils"
	"github.com/tuxdudehomelab/homelab/internal/testhelpers"
)

var deepCopyTests = []struct {
	name  string
	input interface{}
}{
	{
		name:  "Deep Copy - Nil",
		input: nil,
	},
	{
		name:  "Deep Copy - Bool - True",
		input: true,
	},
	{
		name:  "Deep Copy - Bool - False",
		input: false,
	},
	{
		name:  "Deep Copy - Bool Pointer - Nil",
		input: newutils.NewNilBool(),
	},
	{
		name:  "Deep Copy - Bool Pointer - True",
		input: newutils.NewBool(true),
	},
	{
		name:  "Deep Copy - Bool Pointer - False",
		input: newutils.NewBool(false),
	},
	{
		name:  "Deep Copy - Int - Zero Value",
		input: 0,
	},
	{
		name:  "Deep Copy - Int - Non-Zero Value",
		input: 78901234,
	},
	{
		name:  "Deep Copy - Int Pointer - Nil",
		input: newutils.NewNilInt(),
	},
	{
		name:  "Deep Copy - Int Pointer - Zero Value",
		input: newutils.NewInt(0),
	},
	{
		name:  "Deep Copy - Int Pointer - Non-Zero Value",
		input: newutils.NewInt(-62825094),
	},
	{
		name:  "Deep Copy - Int8 - Zero Value",
		input: int8(0),
	},
	{
		name:  "Deep Copy - Int8 - Non-Zero Value",
		input: int8(87),
	},
	{
		name:  "Deep Copy - Int8 Pointer - Nil",
		input: newutils.NewNilInt8(),
	},
	{
		name:  "Deep Copy - Int8 Pointer - Zero Value",
		input: newutils.NewInt8(0),
	},
	{
		name:  "Deep Copy - Int8 Pointer - Non-Zero Value",
		input: newutils.NewInt8(-43),
	},
	{
		name:  "Deep Copy - Int16 - Zero Value",
		input: int16(0),
	},
	{
		name:  "Deep Copy - Int16 - Non-Zero Value",
		input: int16(-3201),
	},
	{
		name:  "Deep Copy - Int16 Pointer - Nil",
		input: newutils.NewNilInt16(),
	},
	{
		name:  "Deep Copy - Int16 Pointer - Zero Value",
		input: newutils.NewInt16(0),
	},
	{
		name:  "Deep Copy - Int16 Pointer - Non-Zero Value",
		input: newutils.NewInt16(7542),
	},
	{
		name:  "Deep Copy - Int32 - Zero Value",
		input: int32(0),
	},
	{
		name:  "Deep Copy - Int32 - Non-Zero Value",
		input: int32(-9321394),
	},
	{
		name:  "Deep Copy - Int32 Pointer - Nil",
		input: newutils.NewNilInt32(),
	},
	{
		name:  "Deep Copy - Int32 Pointer - Zero Value",
		input: newutils.NewInt32(0),
	},
	{
		name:  "Deep Copy - Int32 Pointer - Non-Zero Value",
		input: newutils.NewInt32(-937231951),
	},
	{
		name:  "Deep Copy - Int64 - Zero Value",
		input: int64(0),
	},
	{
		name:  "Deep Copy - Int64 - Non-Zero Value",
		input: int64(24972405272562325),
	},
	{
		name:  "Deep Copy - Int64 Pointer - Nil",
		input: newutils.NewNilInt64(),
	},
	{
		name:  "Deep Copy - Int64 Pointer - Zero Value",
		input: newutils.NewInt64(0),
	},
	{
		name:  "Deep Copy - Int64 Pointer - Non-Zero Value",
		input: newutils.NewInt64(-2978903752356259252),
	},
	{
		name:  "Deep Copy - Uint - Zero Value",
		input: uint(0),
	},
	{
		name:  "Deep Copy - Uint - Non-Zero Value",
		input: uint(352352525252),
	},
	{
		name:  "Deep Copy - Uint Pointer - Nil",
		input: newutils.NewNilUint(),
	},
	{
		name:  "Deep Copy - Uint Pointer - Zero Value",
		input: newutils.NewUint(0),
	},
	{
		name:  "Deep Copy - Uint Pointer - Non-Zero Value",
		input: newutils.NewUint(9357355265295),
	},
	{
		name:  "Deep Copy - Uint8 - Zero Value",
		input: uint8(0),
	},
	{
		name:  "Deep Copy - Uint8 - Non-Zero Value",
		input: uint8(21),
	},
	{
		name:  "Deep Copy - Uint8 Pointer - Nil",
		input: newutils.NewNilUint8(),
	},
	{
		name:  "Deep Copy - Uint8 Pointer - Zero Value",
		input: newutils.NewUint8(0),
	},
	{
		name:  "Deep Copy - Uint8 Pointer - Non-Zero Value",
		input: newutils.NewUint8(243),
	},
	{
		name:  "Deep Copy - Uint16 - Zero Value",
		input: uint16(0),
	},
	{
		name:  "Deep Copy - Uint16 - Non-Zero Value",
		input: uint16(9335),
	},
	{
		name:  "Deep Copy - Uint16 Pointer - Nil",
		input: newutils.NewNilUint16(),
	},
	{
		name:  "Deep Copy - Uint16 Pointer - Zero Value",
		input: newutils.NewUint16(0),
	},
	{
		name:  "Deep Copy - Uint16 Pointer - Non-Zero Value",
		input: newutils.NewUint16(9224),
	},
	{
		name:  "Deep Copy - Uint32 - Zero Value",
		input: uint32(0),
	},
	{
		name:  "Deep Copy - Uint32 - Non-Zero Value",
		input: uint32(845345395),
	},
	{
		name:  "Deep Copy - Uint32 Pointer - Nil",
		input: newutils.NewNilUint32(),
	},
	{
		name:  "Deep Copy - Uint32 Pointer - Zero Value",
		input: newutils.NewUint32(0),
	},
	{
		name:  "Deep Copy - Uint32 Pointer - Non-Zero Value",
		input: newutils.NewUint32(874274872),
	},
	{
		name:  "Deep Copy - Uint64 - Zero Value",
		input: uint64(0),
	},
	{
		name:  "Deep Copy - Uint64 - Non-Zero Value",
		input: uint64(3535356353536363447),
	},
	{
		name:  "Deep Copy - Uint64 Pointer - Nil",
		input: newutils.NewNilUint64(),
	},
	{
		name:  "Deep Copy - Uint64 Pointer - Zero Value",
		input: newutils.NewUint64(0),
	},
	{
		name:  "Deep Copy - Uint64 Pointer - Non-Zero Value",
		input: newutils.NewUint64(975937538563875368),
	},
	{
		name:  "Deep Copy - Uintptr - Zero Value",
		input: uintptr(0),
	},
	{
		name:  "Deep Copy - Uintptr - Non-Zero Value",
		input: uintptr(86463356),
	},
	{
		name:  "Deep Copy - Float32 - Zero Value",
		input: float32(0),
	},
	{
		name:  "Deep Copy - Float32 - Non-Zero Value",
		input: float32(-1.698636335532),
	},
	{
		name:  "Deep Copy - Float32 Pointer - Nil",
		input: newutils.NewNilFloat32(),
	},
	{
		name:  "Deep Copy - Float32 Pointer - Zero Value",
		input: newutils.NewFloat32(0),
	},
	{
		name:  "Deep Copy - Float32 Pointer - Non-Zero Value",
		input: newutils.NewFloat32(95645.552527562),
	},
	{
		name:  "Deep Copy - Float64 - Zero Value",
		input: float64(0),
	},
	{
		name:  "Deep Copy - Float64 - Non-Zero Value",
		input: float64(2424252.56473525),
	},
	{
		name:  "Deep Copy - Float64 Pointer - Nil",
		input: newutils.NewNilFloat64(),
	},
	{
		name:  "Deep Copy - Float64 Pointer - Zero Value",
		input: newutils.NewFloat64(0),
	},
	{
		name:  "Deep Copy - Float64 Pointer - Non-Zero Value",
		input: newutils.NewFloat64(-52525.52636222523),
	},
	{
		name:  "Deep Copy - Complex64 - Zero Value",
		input: complex64(0),
	},
	{
		name:  "Deep Copy - Complex64 - Non-Zero Value",
		input: complex64(-943.5454 + 9353.7463i),
	},
	{
		name:  "Deep Copy - Complex64 Pointer - Nil",
		input: newutils.NewNilComplex64(),
	},
	{
		name:  "Deep Copy - Complex64 Pointer - Zero Value",
		input: newutils.NewComplex64(0),
	},
	{
		name:  "Deep Copy - Complex64 Pointer - Non-Zero Value",
		input: newutils.NewComplex64(535.5353 - 636363.52342i),
	},
	{
		name:  "Deep Copy - Complex128 - Zero Value",
		input: complex128(0),
	},
	{
		name:  "Deep Copy - Complex128 - Non-Zero Value",
		input: complex128(9424242.52552 + 87446453.2324i),
	},
	{
		name:  "Deep Copy - Complex128 Pointer - Nil",
		input: newutils.NewNilComplex128(),
	},
	{
		name:  "Deep Copy - Complex128 Pointer - Zero Value",
		input: newutils.NewComplex128(0),
	},
	{
		name:  "Deep Copy - Complex128 Pointer - Non-Zero Value",
		input: newutils.NewComplex128(-6353453.636363 - 96435334.2425i),
	},
	{
		name:  "Deep Copy - Array - Zero Length",
		input: [0]int{},
	},
	{
		name:  "Deep Copy - Array - Non-Zero Length",
		input: [6]int{-1, 794, 525, -78463, -22, 76},
	},
	{
		name:  "Deep Copy - Map - Zero Value",
		input: map[string]int(nil),
	},
	{
		name: "Deep Copy - Map - Non-Zero Value",
		input: map[string]int{
			"foo": 123,
			"bar": 456,
		},
	},
	{
		name:  "Deep Copy - Slice - Zero Value",
		input: []int(nil),
	},
	{
		name:  "Deep Copy - Slice - Non-Zero Value",
		input: []int{-3535353, 7474, 4242424, -633655, 2242, -73665353, 242424},
	},
	{
		name:  "Deep Copy - Chan - Nil",
		input: chan struct{}(nil),
	},
	{
		name:  "Deep Copy - Interface - Zero Value",
		input: any(nil),
	},
	{
		name:  "Deep Copy - Interface - Recursive - Nil",
		input: any(any(any(any(any(nil))))),
	},
	{
		name:  "Deep Copy - Interface - Recursive - Non-Nil",
		input: any(any(any(any(any([]int{1, 2, 3}))))),
	},
	{
		name:  "Deep Copy - Interface - Scalar",
		input: any(1),
	},
	{
		name:  "Deep Copy - Interface - Array",
		input: any([4]int{1, 2, 3, 4}),
	},
	{
		name:  "Deep Copy - Interface - Slice",
		input: any([]int{1, 2, 3, 4}),
	},
	{
		name:  "Deep Copy - Array of Interfaces",
		input: [1]any{123},
	},
	{
		name:  "Deep Copy - Slice of Interfaces",
		input: []any{"foobar"},
	},
	{
		name:  "Deep Copy - String - Zero Value",
		input: "",
	},
	{
		name:  "Deep Copy - String - Non-Zero Value",
		input: "foobar",
	},
	{
		name:  "Deep Copy - Struct - Zero Value",
		input: struct{}{},
	},
	{
		name:  "Deep Copy - Struct - Unexported",
		input: simpleUnexportedStruct{f1: "foo", f2: 101},
	},
	{
		name:  "Deep Copy - Pointer To Struct - Unexported",
		input: &simpleUnexportedStruct{f1: "foo", f2: 101},
	},
	{
		name: "Deep Copy - Struct - Exhaustive",
		input: struct1{
			F1: "foobar",
			F2: 123,
			f3: true,
			f4: 1234567890123,
			F5: struct2{
				f1: map[int]interface{}{
					12: "foo",
					13: true,
					14: nil,
				},
				f2: []*struct1{
					{
						F1: "xyz",
						f3: true,
					},
					{
						f4: 987654321,
					},
					{
						f7: [3]int{0, 1, 2},
					},
				},
			},
			F6: &struct1{
				f3: true,
				f7: [3]int{-1, 0, 1},
			},
		},
	},
}

func TestDeepCopy(t *testing.T) {
	t.Parallel()

	for _, test := range deepCopyTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, gotErr := Copy(tc.input)
			if gotErr != nil {
				testhelpers.LogErrorNotNil(t, "Copy()", tc.name, gotErr)
				return
			}

			if !reflect.DeepEqual(got, tc.input) {
				testhelpers.LogCustom(t, "Copy()", tc.name, fmt.Sprintf("got %v != want %v", got, tc.input))
				return
			}
		})
	}
}

var deepCopyErrorTests = []struct {
	name  string
	input interface{}
	want  string
}{
	{
		name:  "Deep Copy - Array - Error",
		input: [1]func(int, bool){func(a int, b bool) {}},
		want:  `Deep Copy failed, reason: cannot handle non-nil type func\(int, bool\)`,
	},
	{
		name:  "Deep Copy - Chan - Error",
		input: makeIntChan(),
		want:  `Deep Copy failed, reason: cannot handle non-nil type chan int`,
	},
	{
		name:  "Deep Copy - Func - Error",
		input: func(a int, b string) {},
		want:  `Deep Copy failed, reason: cannot handle non-nil type func\(int, string\)`,
	},
	{
		name: "Deep Copy - Map - Error",
		input: map[string]func(bool){
			"foobar": func(b bool) {},
		},
		want: `Deep Copy failed, reason: cannot handle non-nil type func\(bool\)`,
	},
	{
		name:  "Deep Copy - Pointer - Error",
		input: makeIntChanPtr(),
		want:  `Deep Copy failed, reason: cannot handle non-nil type chan int`,
	},
	{
		name:  "Deep Copy - Slice - Error",
		input: []func(int){func(a int) {}},
		want:  `Deep Copy failed, reason: cannot handle non-nil type func\(int\)`,
	},
	{
		name: "Deep Copy - Struct - Error",
		input: struct2{
			F3: func(a int, b string) bool { return true },
		},
		want: `Deep Copy failed, reason: cannot handle non-nil type func\(int, string\) bool`,
	},
}

func TestDeepCopyErrors(t *testing.T) {
	t.Parallel()

	for _, test := range deepCopyErrorTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, gotErr := Copy(tc.input)
			if gotErr == nil {
				testhelpers.LogErrorNil(t, "Copy()", tc.name, tc.want)
				return
			}

			if !testhelpers.RegexMatch(t, "Copy()", tc.name, "gotErr error string", tc.want, gotErr.Error()) {
				return
			}
		})
	}
}

func TestDeepMustCopyErrors(t *testing.T) {
	t.Parallel()

	for _, test := range deepCopyErrorTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defer testhelpers.ExpectPanic(t, "MustCopy()", tc.name, tc.want)
			_ = MustCopy(tc.input)
		})
	}
}

type struct1 struct {
	F1 string
	F2 int
	f3 bool
	f4 uint64
	F5 struct2
	F6 *struct1
	f7 [3]int
}

type struct2 struct {
	f1 map[int]interface{}
	f2 []*struct1
	F3 func(int, string) bool
	F4 chan struct{}
}

type simpleUnexportedStruct struct {
	f1 string
	f2 int64
}

func makeIntChan() chan int {
	ch := make(chan int)
	return ch
}

func makeIntChanPtr() *chan int {
	ch := make(chan int)
	return &ch
}
