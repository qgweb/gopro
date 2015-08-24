package convert

import (
	"strconv"
)

func ToFloat32(d interface{}) (f float32) {
	switch d.(type) {
	case int:
		f = float32(d.(int))
		return
	case int32:
		f = float32(d.(int32))
		return
	case int64:
		f = float32(d.(int64))
		return
	case float32:
		f = d.(float32)
	case float64:
		f = float32(d.(float64))
		return
	case byte:
		f = float32(d.(byte))
		return
	case string:
		t, _ := strconv.ParseFloat(d.(string), 32)
		f = float32(t)
		return
	}
	return 0.0
}

func ToFloat64(d interface{}) (f float64) {
	switch d.(type) {
	case int:
		f = float64(d.(int))
		return
	case int32:
		f = float64(d.(int32))
		return
	case int64:
		f = float64(d.(int64))
		return
	case float32:
		f = float64(d.(float32))
		return
	case float64:
		f = d.(float64)
		return
	case byte:
		f = float64(d.(byte))
		return
	case string:
		f, _ = strconv.ParseFloat(d.(string), 64)
		return
	}
	return 0.0
}
