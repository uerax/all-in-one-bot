package common

import "reflect"

func InSlice(slice, val any) bool {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return false
	}
	for i := 0; i < v.Len(); i++ {
		if reflect.DeepEqual(val, v.Index(i).Interface()) {
            return true
        }
	}
	return false
}