package mock

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

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
	inOrder           bool
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

type AnyType string
type AnythingOfType string

const (
	Any AnyType = "mock.any"
)

func AnyOfType(t string) AnythingOfType {
	return AnythingOfType(t)
}

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

func (m *Mock) When(name string, arguments ...interface{}) *MockFunction {
	f := &MockFunction{
		Name:      name,
		Arguments: arguments,
	}

	m.Functions = append(m.Functions, f)
	return f
}

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

	panic(fmt.Sprintf("Mock call missing %s (%v)", functionName, arguments))
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

func (f *MockFunction) Return(v ...interface{}) *MockFunction {
	f.ReturnValues = append(f.ReturnValues, v...)
	return f
}

func (f *MockFunction) ReturnToArgument(n int, v interface{}) *MockFunction {
	f.ReturnToArguments = append(f.ReturnToArguments, MockReturnToArgument{n, v})
	return f
}

func (f *MockFunction) Panic(v interface{}) *MockFunction {
	f.PanicValue = v
	return f
}

func (f *MockFunction) InOrder(v bool) *MockFunction {
	panic("FIXME: unsupported")
	f.inOrder = v
	return f
}

func (f *MockFunction) Times(a int) *MockFunction {
	f.countCheck = TIMES
	f.times = [2]int{-1, a}
	return f
}

func (f *MockFunction) AtLeast(a int) *MockFunction {
	f.countCheck = AT_LEAST
	f.times = [2]int{-1, a}
	return f
}

func (f *MockFunction) AtMost(a int) *MockFunction {
	f.countCheck = AT_MOST
	f.times = [2]int{-1, a}
	return f
}

func (f *MockFunction) Between(a, b int) *MockFunction {
	f.countCheck = BETWEEN
	f.times = [2]int{a, b}
	return f
}

func (r *MockResult) Get(i int) interface{} {
	return r.Result[i]
}

func (r *MockResult) Int(i int) int {
	return r.Result[i].(int)
}

func (r *MockResult) Int8(i int) int8 {
	return r.Result[i].(int8)
}

func (r *MockResult) Int16(i int) int16 {
	return r.Result[i].(int16)
}

func (r *MockResult) Int32(i int) int32 {
	return r.Result[i].(int32)
}

func (r *MockResult) Int64(i int) int64 {
	return r.Result[i].(int64)
}

func (r *MockResult) Bool(i int) bool {
	return r.Result[i].(bool)
}

func (r *MockResult) Float32(i int) float32 {
	return r.Result[i].(float32)
}

func (r *MockResult) Float64(i int) float64 {
	return r.Result[i].(float64)
}

func (r *MockResult) String(i int) string {
	return r.Result[i].(string)
}

func (r *MockResult) Error(i int) error {
	if r.Result[i] == nil {
		return nil
	} else {
		return r.Result[i].(error)
	}
}
