package v8tsgo

import (
	"fmt"

	v8 "rogchap.com/v8go"
)

type V8Utils struct {
	ctx *v8.Context
	goUtils *v8.Object
	fnCreateError *v8.Function
}

func NewV8Utils(ctx *v8.Context) (*V8Utils, error) {
	utils := &V8Utils{
		ctx: ctx,
	}
	_, err := ctx.RunScript("var _go_utils = {}; _go_utils.create_error(msg) = (msg) => new Error(msg);", "init_go_utils.js")
	if err != nil {
		return nil, fmt.Errorf("failed to execute the go utils init script, %w", err)
	}
	valUtils, err := ctx.Global().Get("_go_utils")
	if err != nil {
		return nil, fmt.Errorf("unable to access the go utils value, %w", err)
	}
	goUtils, err := valUtils.AsObject()
	if err != nil {
		return nil, fmt.Errorf("unable to cast the go utils value to an object, %w", err)
	}
	utils.goUtils = goUtils
	valCreateError, err := goUtils.Get("create_error")
	if err != nil {
		return nil, fmt.Errorf("unable to access the create_error value, %w", err)
	}
	fnCreateError, err := valCreateError.AsFunction()
	if err != nil {
		return nil, fmt.Errorf("unable to cast the create_error value to an function, %w", err)
	}
	utils.fnCreateError = fnCreateError
	return utils, nil
}

func (u *V8Utils) WrapError(err error) (*v8.Value, error) {
	valMsg, err := v8.NewValue(u.ctx.Isolate(), err.Error())
	if err != nil {
		return nil, fmt.Errorf("unable to create msg value, %w", err)
	}
	return u.fnCreateError.Call(u.goUtils, valMsg)
}