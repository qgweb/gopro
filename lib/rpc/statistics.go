package rpc

import (
	"reflect"

	"github.com/hprose/hprose-go/hprose"
)

type MyServiceEvent struct {
	OnBeforeInvokeFun func(string, []reflect.Value, bool, hprose.Context)
	OnAfterInvokeFun  func(string, []reflect.Value, bool, []reflect.Value, hprose.Context)
	OnSendErrorFun    func(error, hprose.Context)
}

func (this *MyServiceEvent) OnBeforeInvoke(name string, args []reflect.Value, byref bool, context hprose.Context) {
	if this.OnBeforeInvokeFun != nil {
		this.OnBeforeInvokeFun(name, args, byref, context)
	}
}

func (this *MyServiceEvent) OnAfterInvoke(name string, args []reflect.Value, byref bool, result []reflect.Value, context hprose.Context) {
	if this.OnAfterInvokeFun != nil {
		this.OnAfterInvokeFun(name, args, byref, result, context)
	}
}

func (this *MyServiceEvent) OnSendError(err error, context hprose.Context) {
	if this.OnSendErrorFun != nil {
		this.OnSendErrorFun(err, context)
	}
}

func (this *MyServiceEvent) OnSendHeader(c *hprose.HttpContext) {
}
