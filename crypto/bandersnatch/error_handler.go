package bandersnatch

import "sync"

type InvalidPointErrorHandler func(reason string, comparison bool, points ...CurvePointRead) bool

func trivial_error_handler(string, bool, ...CurvePointRead) bool {
	return false
}

func panic_error_handler(reason string, _ bool, _ ...CurvePointRead) bool {
	panic(reason)
}

func wasInvalidPointEncountered(fun func()) bool {
	var old_handler_ptr *InvalidPointErrorHandler // indirection to avoid locking.
	var error_bit bool = false
	new_handler := func(reason string, comparison bool, points ...CurvePointRead) bool {
		error_bit = true
		return (*old_handler_ptr)(reason, comparison, points...)
	}
	old_handler := SetInvalidPointErrorHandler(new_handler)
	old_handler_ptr = &old_handler
	defer SetInvalidPointErrorHandler(*old_handler_ptr)
	fun()
	return error_bit
}

var current_error_handler InvalidPointErrorHandler = trivial_error_handler
var error_handler_mutex sync.Mutex

func SetInvalidPointErrorHandler(new_handler InvalidPointErrorHandler) (old_handler InvalidPointErrorHandler) {
	error_handler_mutex.Lock()
	defer error_handler_mutex.Unlock()
	old_handler = current_error_handler
	current_error_handler = new_handler
	return
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
