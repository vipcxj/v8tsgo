package v8tsgo

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"
	"time"

	"rogchap.com/v8go"
	"github.com/vipcxj/v8tsgo/internal/test"
)

type testObject struct {
	A int
	B string
	C *float32
	D struct {
		E bool
		F []int
	}
}

func isObjectOrArrayEquals(ctx *v8go.Context, a any, b any) bool {
	var err error
	var ma any
	var mb any
	var ja []byte
	var jb []byte
	switch v := a.(type) {
	case v8go.Valuer:
		sa, err := v8go.JSONStringify(ctx, v)
		panicIfErr(err)
		ja = []byte(sa)
	default:
		ja, err = json.Marshal(a)
		panicIfErr(err)
	}
	switch v := b.(type) {
	case v8go.Valuer:
		sb, err := v8go.JSONStringify(ctx, v)
		panicIfErr(err)
		jb = []byte(sb)
	default:
		jb, err = json.Marshal(b)
		panicIfErr(err)
	}
	err = json.Unmarshal(ja, &ma)
	panicIfErr(err)
	err = json.Unmarshal(jb, &mb)
	panicIfErr(err)
	return reflect.DeepEqual(ma, mb)
}

func TestMakeValue(t *testing.T) {
	ctx := v8go.NewContext()
	var null any
	v, err := MakeValue(ctx, null)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsNull(), "")

	var ptrInt *int = nil
	null = ptrInt
	v, err = MakeValue(ctx, null)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsNull(), "")

	var ptrTo *testObject = nil
	null = ptrTo
	v, err = MakeValue(ctx, null)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsNull(), "")

	b := true
	v, err = MakeValue(ctx, b)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsBoolean(), "")
	test.AssertEqual(t, true, v.Boolean(), "")
	v, err = MakeValue(ctx, &b)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsBoolean(), "")
	test.AssertEqual(t, true, v.Boolean(), "")

	s := "abc"
	v, err = MakeValue(ctx, s)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsString(), "")
	test.AssertEqual(t, "abc", v.String(), "")
	v, err = MakeValue(ctx, &s)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsString(), "")
	test.AssertEqual(t, "abc", v.String(), "")

	var ui8 uint8 = 1
	v, err = MakeValue(ctx, ui8)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsUint32(), "")
	test.AssertEqual(t, 1, v.Uint32(), "")
	v, err = MakeValue(ctx, &ui8)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsUint32(), "")
	test.AssertEqual(t, 1, v.Uint32(), "")

	var ui16 uint16 = 10
	v, err = MakeValue(ctx, ui16)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsUint32(), "")
	test.AssertEqual(t, 10, v.Uint32(), "")
	v, err = MakeValue(ctx, &ui16)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsUint32(), "")
	test.AssertEqual(t, 10, v.Uint32(), "")

	var ui32 uint32 = 100
	v, err = MakeValue(ctx, ui32)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsUint32(), "")
	test.AssertEqual(t, 100, v.Uint32(), "")
	v, err = MakeValue(ctx, &ui32)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsUint32(), "")
	test.AssertEqual(t, 100, v.Uint32(), "")

	var ui64 uint64 = 1000
	v, err = MakeValue(ctx, ui64)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsBigInt(), "")
	test.AssertEqual(t, 1000, v.BigInt().Uint64(), "")
	v, err = MakeValue(ctx, &ui64)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsBigInt(), "")
	test.AssertEqual(t, 1000, v.BigInt().Uint64(), "")

	var i8 int8 = 1
	v, err = MakeValue(ctx, i8)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsInt32(), "")
	test.AssertEqual(t, 1, v.Int32(), "")
	v, err = MakeValue(ctx, &i8)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsInt32(), "")
	test.AssertEqual(t, 1, v.Int32(), "")

	var i16 int16 = 10
	v, err = MakeValue(ctx, i16)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsInt32(), "")
	test.AssertEqual(t, 10, v.Int32(), "")
	v, err = MakeValue(ctx, &i16)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsInt32(), "")
	test.AssertEqual(t, 10, v.Int32(), "")

	var i32 int32 = 100
	v, err = MakeValue(ctx, i32)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsInt32(), "")
	test.AssertEqual(t, 100, v.Int32(), "")
	v, err = MakeValue(ctx, &i32)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsInt32(), "")
	test.AssertEqual(t, 100, v.Int32(), "")

	var i64 int64 = 1000
	v, err = MakeValue(ctx, i64)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsBigInt(), "")
	test.AssertEqual(t, 1000, v.BigInt().Int64(), "")
	v, err = MakeValue(ctx, &i64)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsBigInt(), "")
	test.AssertEqual(t, 1000, v.BigInt().Int64(), "")

	var i int = 10000
	v, err = MakeValue(ctx, i)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsUint32(), "")
	test.AssertEqual(t, 10000, v.Integer(), "")
	v, err = MakeValue(ctx, &i)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsUint32(), "")
	test.AssertEqual(t, 10000, v.Integer(), "")

	i = -i
	v, err = MakeValue(ctx, i)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsInt32(), "")
	test.AssertEqual(t, int64(i), v.Integer(), "")
	v, err = MakeValue(ctx, &i)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsInt32(), "")
	test.AssertEqual(t, int64(i), v.Integer(), "")

	i = math.MaxUint32 + 1
	v, err = MakeValue(ctx, i)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsBigInt(), "")
	test.AssertEqual(t, int64(i), v.BigInt().Int64(), "")
	v, err = MakeValue(ctx, &i)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsBigInt(), "")
	test.AssertEqual(t, int64(i), v.BigInt().Int64(), "")

	i = math.MinInt32 - 1
	v, err = MakeValue(ctx, i)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsBigInt(), "")
	test.AssertEqual(t, int64(i), v.BigInt().Int64(), "")
	v, err = MakeValue(ctx, &i)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsBigInt(), "")
	test.AssertEqual(t, int64(i), v.BigInt().Int64(), "")

	var f32 float32 = 3.14
	v, err = MakeValue(ctx, f32)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsNumber(), "")
	test.AssertEqual(t, float64(f32), v.Number(), "")
	v, err = MakeValue(ctx, &f32)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsNumber(), "")
	test.AssertEqual(t, float64(f32), v.Number(), "")

	var f64 float64 = 3.14159265
	v, err = MakeValue(ctx, f64)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsNumber(), "")
	test.AssertEqual(t, f64, v.Number(), "")
	v, err = MakeValue(ctx, &f64)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsNumber(), "")
	test.AssertEqual(t, f64, v.Number(), "")

	var ts time.Time = time.Now()
	v, err = MakeValue(ctx, ts)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsDate(), "")
	tsV, err := dateGetTime(v)
	panicIfErr(err)
	test.AssertEqual(t, ts.UnixMilli(), tsV, "")
	v, err = MakeValue(ctx, &ts)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsDate(), "")
	tsV, err = dateGetTime(v)
	panicIfErr(err)
	test.AssertEqual(t, ts.UnixMilli(), tsV, "")

	o := testObject {
		A: 1,
		B: "2",
		C: &f32,
		D: struct{E bool; F []int}{
			E: true,
			F: []int { 7, 8, 9 },
		},
	}
	v, err = MakeValue(ctx, o)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsObject(), "")
	test.AssertEqual(t, true, isObjectOrArrayEquals(ctx, o, v), "")
	v, err = MakeValue(ctx, &o)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsObject(), "")
	test.AssertEqual(t, true, isObjectOrArrayEquals(ctx, o, v), "")

	sli := []any {
		o, "123", 1, 2.3, true, nil,
	}
	v, err = MakeValue(ctx, sli)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsArray(), "")
	test.AssertEqual(t, true, isObjectOrArrayEquals(ctx, sli, v), "")
	v, err = MakeValue(ctx, &sli)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsArray(), "")
	test.AssertEqual(t, true, isObjectOrArrayEquals(ctx, sli, v), "")

	ar := [6]any {
		o, "123", 1, 2.3, true, nil,
	}
	v, err = MakeValue(ctx, ar)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsArray(), "")
	test.AssertEqual(t, true, isObjectOrArrayEquals(ctx, ar, v), "")
	v, err = MakeValue(ctx, &ar)
	panicIfErr(err)
	test.MustEqual(t, true, v.IsArray(), "")
	test.AssertEqual(t, true, isObjectOrArrayEquals(ctx, ar, v), "")
}

func TestParseValue(t *testing.T) {
	ctx := v8go.NewContext()
	iso := ctx.Isolate()
	v, err := v8go.NewValue(iso, "abc")
	panicIfErr(err)
	var str string
	panicIfErr(ParseValue(ctx, v, &str))
	test.AssertEqual(t, "abc", str, "")

	v, err = v8go.NewValue(iso, true)
	panicIfErr(err)
	var bVar bool
	panicIfErr(ParseValue(ctx, v, &bVar))
	test.AssertEqual(t, true, bVar, "")

	now := time.Now()
	v, err = MakeValue(ctx, now)
	panicIfErr(err)
	var ts time.Time
	panicIfErr(ParseValue(ctx, v, &ts))
	test.AssertEqual(t, now.UnixMilli(), ts.UnixMilli(), "")
}