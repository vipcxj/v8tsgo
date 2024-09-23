package v8tsgo

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"time"

	"rogchap.com/v8go"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func MakeValue(ctx *v8go.Context, goVal any) (*v8go.Value, error) {
	iso := ctx.Isolate()
	if goVal == nil {
		return v8go.Null(iso), nil
	}
	switch v := goVal.(type) {
	case *v8go.Value:
		return v, nil
	case string:
		return v8go.NewValue(iso, v)
	case bool:
		return v8go.NewValue(iso, v)
	case int:
		if v >= 0 && v <= math.MaxUint32 {
			return v8go.NewValue(iso, uint32(v))
		} else if v < 0 && v >= math.MinInt32 {
			return v8go.NewValue(iso, int32(v))
		} else {
			return v8go.NewValue(iso, int64(v))
		}
	case int8:
		return v8go.NewValue(iso, int32(v))
	case int16:
		return v8go.NewValue(iso, int32(v))
	case int32:
		return v8go.NewValue(iso, v)
	case int64:
		return v8go.NewValue(iso, v)
	case uint8:
		return v8go.NewValue(iso, uint32(v))
	case uint16:
		return v8go.NewValue(iso, uint32(v))
	case uint32:
		return v8go.NewValue(iso, v)
	case uint64:
		return v8go.NewValue(iso, v)
	case float32:
		return v8go.NewValue(iso, float64(v))
	case float64:
		return v8go.NewValue(iso, v)
	case time.Time:
		return ctx.RunScript(fmt.Sprintf("new Date(%d)", v.UnixMilli()), fmt.Sprintf("create-date-%d.js", v.UnixMilli()))
	case *string:
		if v == nil {
			return v8go.Null(iso), nil
		}
		return v8go.NewValue(iso, *v)
	case *bool:
		if v == nil {
			return v8go.Null(iso), nil
		}
		return v8go.NewValue(iso, *v)
	case *int:
		if v == nil {
			return v8go.Null(iso), nil
		}
		vv := *v
		if vv >= 0 && vv <= math.MaxUint32 {
			return v8go.NewValue(iso, uint32(vv))
		} else if vv < 0 && vv >= math.MinInt32 {
			return v8go.NewValue(iso, int32(vv))
		} else {
			return v8go.NewValue(iso, int64(vv))
		}
	case *int8:
		if v == nil {
			return v8go.Null(iso), nil
		}
		return v8go.NewValue(iso, int32(*v))
	case *int16:
		if v == nil {
			return v8go.Null(iso), nil
		}
		return v8go.NewValue(iso, int32(*v))
	case *int32:
		if v == nil {
			return v8go.Null(iso), nil
		}
		return v8go.NewValue(iso, *v)
	case *int64:
		if v == nil {
			return v8go.Null(iso), nil
		}
		return v8go.NewValue(iso, *v)
	case *uint8:
		if v == nil {
			return v8go.Null(iso), nil
		}
		return v8go.NewValue(iso, uint32(*v))
	case *uint16:
		if v == nil {
			return v8go.Null(iso), nil
		}
		return v8go.NewValue(iso, uint32(*v))
	case *uint32:
		if v == nil {
			return v8go.Null(iso), nil
		}
		return v8go.NewValue(iso, *v)
	case *uint64:
		if v == nil {
			return v8go.Null(iso), nil
		}
		return v8go.NewValue(iso, *v)
	case *float32:
		if v == nil {
			return v8go.Null(iso), nil
		}
		return v8go.NewValue(iso, float64(*v))
	case *float64:
		if v == nil {
			return v8go.Null(iso), nil
		}
		return v8go.NewValue(iso, *v)
	case *big.Int:
		if v == nil {
			return v8go.Null(iso), nil
		}
		return v8go.NewValue(iso, v)
	case *time.Time:
		return ctx.RunScript(fmt.Sprintf("new Date(%d)", v.UnixMilli()), fmt.Sprintf("create-date-%d.js", v.UnixMilli()))
	case error:
		errStr := v.Error()
		errBytes, err := json.Marshal(errStr)
		if err != nil {
			return nil, fmt.Errorf("unable to create v8 Error object from go error \"%s\", %w", errStr, err)
		}
		return ctx.RunScript(fmt.Sprintf("new Error(%s)", string(errBytes)), fmt.Sprintf("create-error-%s.js", errStr))
	default:
		rv := reflect.ValueOf(v)
		for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
			rv = rv.Elem()
		}
		switch rv.Kind() {
		case reflect.Struct:
			fallthrough
		case reflect.Slice:
			fallthrough
		case reflect.Array:
			j, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("unable to marshal %v to json when making a v8 value, %w", v, err)
			}
			return v8go.JSONParse(ctx, string(j))
		case reflect.Uint8:
			fallthrough
		case reflect.Uint16:
			fallthrough
		case reflect.Uint32:
			return v8go.NewValue(iso, uint32(rv.Uint()))
		case reflect.Uint64:
			return v8go.NewValue(iso, rv.Uint())
		case reflect.Int8:
			fallthrough
		case reflect.Int16:
			fallthrough
		case reflect.Int32:
			return v8go.NewValue(iso, int32(rv.Uint()))
		case reflect.Int64:
			return v8go.NewValue(iso, rv.Int())
		case reflect.Int:
			v := rv.Int()
			if v >= 0 && v <= math.MaxUint32 {
				return v8go.NewValue(iso, uint32(v))
			} else if v < 0 && v >= math.MinInt32 {
				return v8go.NewValue(iso, int32(v))
			} else {
				return v8go.NewValue(iso, int64(v))
			}
		case reflect.Float32:
			fallthrough
		case reflect.Float64:
			return v8go.NewValue(iso, rv.Float())
		case reflect.Bool:
			return v8go.NewValue(iso, rv.Bool())
		case reflect.String:
			return v8go.NewValue(iso, rv.String())
		case reflect.Invalid:
			return v8go.Null(iso), nil
		default:
			return v8go.NewValue(iso, v)
		}
	}
}

func dateGetTime(value *v8go.Value) (int64, error) {
	tsValue, err := value.Object().MethodCall("getTime")
	if err != nil {
		return 0, err
	}
	return tsValue.Integer(), nil
}

func ParseValue(ctx *v8go.Context, value *v8go.Value, out any) error {
	if value.IsNull() || value.IsUndefined() {
		return nil
	}
	switch o := out.(type) {
	case *bool:
		if !value.IsBoolean() {
			return fmt.Errorf("the input value is not a boolean value")
		}
		*o = value.Boolean()
	case *string:
		if !value.IsString() && !value.IsStringObject() {
			return fmt.Errorf("the input value is not a string value")
		}
		*o = value.String()
	case *time.Time:
		if !value.IsDate() {
			return fmt.Errorf("the input value is not a date value")
		}
		ts, err := dateGetTime(value)
		if err != nil {
			return fmt.Errorf("failed to get the timestamp from the v8 date value, %w", err)
		}
		*o = time.UnixMilli(ts)
	default:
		js, err := v8go.JSONStringify(ctx, value)
		if err != nil {
			return fmt.Errorf("unable to json stringify the v8 value, %w", err)
		}
		err = json.Unmarshal([]byte(js), out)
		if err != nil {
			return fmt.Errorf("failed to unmarshal the json string \"%s\", %w", js, err)
		}
	}
	return nil
}