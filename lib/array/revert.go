package array

import "reflect"

func Revert(ary interface{}) {
	r := reflect.ValueOf(ary)

	if r.Type().Kind() != reflect.Ptr || r.Elem().Kind() != reflect.Slice {
		panic("参数必须是切片类型")
	}

	r = r.Elem()
	l := r.Len()

	nary := reflect.MakeSlice(r.Type(), l, r.Cap())

	for i := 0; i < l; i++ {
		nary.Index(i).Set(r.Index(l - i - 1))
	}

	r.Set(nary)
}
