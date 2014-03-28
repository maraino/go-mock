package mock

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/kr/pretty"
)

// Mock should be embedded in the struct that we want to act as a Mock.
//
// Example:
// 		type MyClient struct {
//			mock.Mock
// 		}
type Mock struct {
	Functions []*MockFunction
}

type MockCountCheckType int

const (
	NONE MockCountCheckType = iota
	TIMES
	AT_LEAST
	AT_MOST
	BETWEEN
)

type MockFunction struct {
	Name              string
	Arguments         []interface{}
	ReturnValues      []interface{}
	ReturnToArguments []MockReturnToArgument
	PanicValue        interface{}
	count             int
	countCheck        MockCountCheckType
	times             [2]int
}

type MockReturnToArgument struct {
	Argument int
	Value    interface{}
}

type MockResult struct {
	Result []interface{}
}

// The constant Any can be used to replace any kind of argument in the stub configuration.
type AnyType string

const (
	Any AnyType = "mock.any"
)

type AnythingOfType string

// AnyOfType can be used to replace any kind of argument of an specific type in the stub configuration.
func AnyOfType(t string) AnythingOfType {
	return AnythingOfType(t)
}

// Verifies the restrictions set in the stubbing.
func (m *Mock) Verify() (bool, error) {
	for _, f := range m.Functions {
		switch f.countCheck {
		case TIMES:
			if f.count != f.times[1] {
				return false, fmt.Errorf("Function %s executed %d times, expected: %d", f.Name, f.count, f.times[1])
			}
		case AT_LEAST:
			if f.count < f.times[1] {
				return false, fmt.Errorf("Function %s executed %d times, expected at least: %d", f.Name, f.count, f.times[1])
			}
		case AT_MOST:
			if f.count > f.times[1] {
				return false, fmt.Errorf("Function %s executed %d times, expected at most: %d", f.Name, f.count, f.times[1])
			}
		case BETWEEN:
			if f.count < f.times[0] || f.count > f.times[1] {
				return false, fmt.Errorf("Function %s executed %d times, expected between: [%d, %d]", f.Name, f.count, f.times[0], f.times[1])
			}
		}
	}
	return true, nil
}

// Defines an stub of one method with some specific arguments. It returns a *MockFunction
// that can be configured with Return, ReturnToArgument, Panic, ...
func (m *Mock) When(name string, arguments ...interface{}) *MockFunction {
	f := &MockFunction{
		Name:      name,
		Arguments: arguments,
	}

	m.Functions = append(m.Functions, f)
	return f
}

// Used in the mocks to replace the actual task.
//
// Example:
// 		func (m *MyClient) Request(url string) (int, string, error) {
// 			r := m.Called(url)
//			return r.Int(0), r.String(1), r.Error(2)
// 		}
func (m *Mock) Called(arguments ...interface{}) *MockResult {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("Couldn't get the caller information")
	}

	functionPath := runtime.FuncForPC(pc).Name()
	parts := strings.Split(functionPath, ".")
	functionName := parts[len(parts)-1]

	if f := m.find(functionName, arguments...); f != nil {
		// Increase the counter
		f.count++

		if f.PanicValue != nil {
			panic(f.PanicValue)
		}

		// Return to arguments
		for _, r := range f.ReturnToArguments {
			arg := arguments[r.Argument]
			argTyp := reflect.TypeOf(arg)
			argElem := reflect.ValueOf(arg).Elem()
			typ := reflect.TypeOf(r.Value)
			if typ.Kind() == reflect.Ptr {
				if typ == argTyp {
					// *type vs *type
					argElem.Set(reflect.ValueOf(r.Value).Elem())
				} else {
					// *type vs **type
					argElem.Set(reflect.ValueOf(r.Value))
				}
			} else {
				if typ == argTyp.Elem() {
					// type vs *type
					argElem.Set(reflect.ValueOf(r.Value))
				} else {
					// type vs **type
					value := reflect.New(typ).Elem()
					value.Set(reflect.ValueOf(r.Value))
					argElem.Set(value.Addr())
				}
			}
		}

		return &MockResult{f.ReturnValues}
	}

	argsStr := pretty.Sprintf("%# v", arguments)
	argsStr = argsStr[15 : len(argsStr)-1]
	panic(fmt.Sprintf("Mock call missing for %s(%s)", functionName, argsStr))
}

func (m *Mock) find(name string, arguments ...interface{}) *MockFunction {
	for _, f := range m.Functions {
		if f.Name != name {
			continue
		}

		if len(f.Arguments) != len(arguments) {
			continue
		}

		found := true
		for i, arg := range f.Arguments {
			switch arg.(type) {
			case AnyType:
				continue
			case AnythingOfType:
				if string(arg.(AnythingOfType)) == reflect.TypeOf(arguments[i]).String() {
					continue
				} else {
					found = false
				}
			default:
				if reflect.DeepEqual(arg, arguments[i]) || reflect.ValueOf(arg) == reflect.ValueOf(arguments[i]) {
					continue
				} else {
					found = false
				}
			}
		}

		if !found {
			continue
		}

		return f
	}

	return nil
}

// Defines the return values of a *MockFunction.
func (f *MockFunction) Return(v ...interface{}) *MockFunction {
	f.ReturnValues = append(f.ReturnValues, v...)
	return f
}

// Defines the values returned to a specific argument of a *MockFunction.
func (f *MockFunction) ReturnToArgument(n int, v interface{}) *MockFunction {
	f.ReturnToArguments = append(f.ReturnToArguments, MockReturnToArgument{n, v})
	return f
}

// Defines a panic for a specific *MockFunction.
func (f *MockFunction) Panic(v interface{}) *MockFunction {
	f.PanicValue = v
	return f
}

// Defines how many times a *MockFunction must be called.
// This is verified if mock.Verify is called.
func (f *MockFunction) Times(a int) *MockFunction {
	f.countCheck = TIMES
	f.times = [2]int{-1, a}
	return f
}

// Defines the number of times that a *MockFunction must be at least called.
// This is verified if mock.Verify is called.
func (f *MockFunction) AtLeast(a int) *MockFunction {
	f.countCheck = AT_LEAST
	f.times = [2]int{-1, a}
	return f
}

// Defines the number of times that a *MockFunction must be at most called.
// This is verified if mock.Verify is called.
func (f *MockFunction) AtMost(a int) *MockFunction {
	f.countCheck = AT_MOST
	f.times = [2]int{-1, a}
	return f
}

// Defines a range of times that a *MockFunction must be called.
// This is verified if mock.Verify is called.
func (f *MockFunction) Between(a, b int) *MockFunction {
	f.countCheck = BETWEEN
	f.times = [2]int{a, b}
	return f
}

// Returns a specific return parameter.
func (r *MockResult) get(i int) (bool, interface{}) {
	if i >= len(r.Result) {
		return false, nil
	} else {
		return true, r.Result[i]
	}
}

// Returns true if the results have the index i, false otherwise.
func (r *MockResult) Contains(i int) bool {
	if len(r.Result) > i {
		return true
	} else {
		return false
	}
}

// Returns a specific return parameter.
// If a result has not been set, it returns nil,
func (r *MockResult) Get(i int) interface{} {
	if r.Contains(i) {
		return r.Result[i]
	} else {
		return nil
	}
}

// Returns a specific return parameter as a bool.
// If a result has not been set, it returns false.
func (r *MockResult) Bool(i int) bool {
	if r.Contains(i) {
		return r.Result[i].(bool)
	} else {
		return false
	}
}

// Returns a specific return parameter as an error.
// If a result has not been set, it returns nil.
func (r *MockResult) Error(i int) error {
	if r.Contains(i) && r.Result[i] != nil {
		return r.Result[i].(error)
	} else {
		return nil
	}
}

// Returns a specific return parameter as a float32.
// If a result has not been set, it returns 0.
func (r *MockResult) Float32(i int) float32 {
	if r.Contains(i) {
		return r.Result[i].(float32)
	} else {
		return 0
	}
}

// Returns a specific return parameter as a float64.
// If a result has not been set, it returns 0.
func (r *MockResult) Float64(i int) float64 {
	if r.Contains(i) {
		return r.Result[i].(float64)
	} else {
		return 0
	}
}

// Returns a specific return parameter as an int.
// If a result has not been set, it returns 0.
func (r *MockResult) Int(i int) int {
	if r.Contains(i) {
		return r.Result[i].(int)
	} else {
		return 0
	}
}

// Returns a specific return parameter as an int8.
// If a result has not been set, it returns 0.
func (r *MockResult) Int8(i int) int8 {
	if r.Contains(i) {
		return r.Result[i].(int8)
	} else {
		return 0
	}
}

// Returns a specific return parameter as an int16.
// If a result has not been set, it returns 0.
func (r *MockResult) Int16(i int) int16 {
	if r.Contains(i) {
		return r.Result[i].(int16)
	} else {
		return 0
	}
}

// Returns a specific return parameter as an int32.
// If a result has not been set, it returns 0.
func (r *MockResult) Int32(i int) int32 {
	if r.Contains(i) {
		return r.Result[i].(int32)
	} else {
		return 0
	}
}

// Returns a specific return parameter as an int64.
// If a result has not been set, it returns 0.
func (r *MockResult) Int64(i int) int64 {
	if r.Contains(i) {
		return r.Result[i].(int64)
	} else {
		return 0
	}
}

// Returns a specific return parameter as a string.
// If a result has not been set, it returns "".
func (r *MockResult) String(i int) string {
	if r.Contains(i) {
		return r.Result[i].(string)
	} else {
		return ""
	}
}
