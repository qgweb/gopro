package main

import (
	"encoding/json"
)

//json返回
func jsonReturn(data interface{}, err error) []byte {
	type Ret struct {
		Ret  int
		Msg  string
		Data interface{}
	}
	if err != nil {
		d, _ := json.Marshal(&Ret{Ret: 1, Msg: err.Error(), Data: nil})
		return d
	}

	d, _ := json.Marshal(&Ret{Ret: 0, Msg: "", Data: data})
	return d
}
