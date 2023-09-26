package internal

import (
	"reflect"
	"runtime"
)

var SectionKey = struct{}{}
var AutomaticSectionKey = struct{}{}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
