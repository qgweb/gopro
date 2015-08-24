package convert

import (
	"strconv"
)

func ToInt(d interface{}) (v int) {
	switch d.(type) {
	case string:
		v, _ = strconv.Atoi(d.(string))
		return
	case byte:
		v = int(d.(byte))
		return
	case int:
		v = d.(int)
		return
	case int8:
		v = int(d.(int8))
		return v
	case int64:
		v = int(d.(int64))
		return v
	case int32:
		v = int(d.(int32))
		return v
	case float32:
		v = int(d.(float32))
		return
	case float64:
		v = int(d.(float64))
		return

	}
	return 0
}

func ToInt64(d interface{}) (v int64) {
	switch d.(type) {
	case string:
		t, _ := strconv.Atoi(d.(string))
		v = int64(t)
		return
	case byte:
		v = int64(d.(byte))
		return
	case int8:
		v = int64(d.(int8))
		return v
	case int64:
		v = int64(d.(int64))
		return v
	case int32:
		v = int64(d.(int32))
		return v
	case float32:
		v = int64(d.(float32))
		return
	case float64:
		v = int64(d.(float64))
		return

	}
	return 0
}

func ToInt32(d interface{}) (v int32) {
	switch d.(type) {
	case string:
		t, _ := strconv.Atoi(d.(string))
		v = int32(t)
		return
	case byte:
		v = int32(d.(byte))
		return
	case int8:
		v = int32(d.(int8))
		return v
	case int64:
		v = int32(d.(int64))
		return v
	case int32:
		v = int32(d.(int32))
		return v
	case float32:
		v = int32(d.(float32))
		return
	case float64:
		v = int32(d.(float64))
		return

	}
	return 0
}
