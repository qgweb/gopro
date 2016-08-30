package convert

import (
	"strconv"
	"strings"
)

func ToString(d interface{}) (s string) {
	switch d.(type) {
	case int:
		s = strconv.Itoa(d.(int))
		return
	case int32:
		s = strconv.FormatInt(int64(d.(int32)), 10)
		return
	case int64:
		s = strconv.FormatInt(d.(int64), 10)
		return
	case bool:
		s = strconv.FormatBool(d.(bool))
		return
	case float32:
		s = strconv.FormatFloat(float64(d.(float32)), 'f', -1, 32)
		return
	case float64:
		s = strconv.FormatFloat(d.(float64), 'f', -1, 64)
		return
	case byte:
		s = strconv.FormatInt(int64(d.(byte)), 10)
		return
	case []byte:
		return string(d.([]byte))
	case string:
		s = d.(string)
		return
	case []interface{}:
		ss := make([]string, 0, 10)
		for _, v := range d.([]interface{}) {
			if vv, ok := v.(string); ok {
				ss = append(ss, vv)
			}
		}
		s = strings.Join(ss, ",")
		return
	}
	return ""
}
