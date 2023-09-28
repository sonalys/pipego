package internal

import (
	"reflect"
	"runtime"
)

type key int

var SectionKey = key(0)
var AutomaticSectionKey = key(1)

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
