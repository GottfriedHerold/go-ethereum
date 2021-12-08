package bandersnatch

import "sync"

type InvalidPointErrorHandler func(reason string, comparison bool, points ...CurvePointRead) bool

func trivial_error_handler(string, bool, ...CurvePointRead) bool {
	return false
}

func panic_error_handler(reason string, _ bool, _ ...CurvePointRead) bool {
	panic(reason)
}

func expectError(fun func()) bool {
	old_handler := GetInvalidPointErrorHandler()
	var error_bit bool = false
	new_handler := func(reason string, comparison bool, points ...CurvePointRead) bool {
		error_bit = true
		return old_handler(reason, comparison, points...)
	}
	SetInvalidPointErrorHandler(new_handler)
	defer SetInvalidPointErrorHandler(old_handler)
	fun()
	return error_bit
}

var current_error_handler InvalidPointErrorHandler = trivial_error_handler
var error_handler_mutex sync.Mutex

func SetInvalidPointErrorHandler(fun InvalidPointErrorHandler) {
	error_handler_mutex.Lock()
	defer error_handler_mutex.Unlock()
	current_error_handler = fun
}

func GetInvalidPointErrorHandler() InvalidPointErrorHandler {
	error_handler_mutex.Lock()
	defer error_handler_mutex.Unlock()
	f := current_error_handler
	return f
}

func handle_errors(reason string, comparison bool, points ...CurvePointRead) bool {
	f := GetInvalidPointErrorHandler()
	return f(reason, comparison, points...)
}
