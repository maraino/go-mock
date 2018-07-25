package mock

import (
	"testing"
	"fmt"
	"strings"
)

type arg struct {
	V string
}

func (a *arg) String() string {
	return fmt.Sprintf("anArg: %s", a.V)
}

type myMock struct {
	Mock
}

func (m *myMock) SomeFunc(a arg) {
	m.Called(a)
}

func (m *myMock) HasTwoArgs(a arg, b *arg) {
	m.Called(a, b)
}

func TestMessageFromAnyIf(t *testing.T) {
	m := &myMock{}

	m.When("SomeFunc", AnyIf("a struct with foo", func(i interface{}) bool {
		a, ok := i.(arg)
		return ok && a.V == "foo"
	}))

	defer assertMessageContains(t, "a struct with foo")
	m.SomeFunc(arg { "bar"})

}

func TestMessageFromArg(t *testing.T) {
	m := &myMock{}

	m.When("SomeFunc", arg { "foo"})

	defer assertMessageContains(t, "SomeFunc({V:foo})")
	m.SomeFunc(arg{"bar"})

}

func TestMessageFromAnyOfType(t *testing.T) {
	m := &myMock{}

	m.When("SomeFunc", AnyOfType("string"))

	defer assertMessageContains(t, "SomeFunc(string)")
	m.SomeFunc(arg{"bar"})

}

func TestMessageForAnyAndType(t *testing.T) {
	m := &myMock{}

	m.When("HasTwoArgs", Any, AnyOfType("string"))

	defer assertMessageContains(t, "HasTwoArgs(mock.any, string)")
	m.HasTwoArgs(arg{"bar"}, &arg{"baz"})

}

func TestMessageForAnyAndNil(t *testing.T) {
	m := &myMock{}

	m.When("HasTwoArgs", Any, nil)

	defer assertMessageContains(t, "HasTwoArgs(mock.any, <nil>)")
	m.HasTwoArgs(arg{"bar"}, &arg{"bar"})

}

func assertMessageContains(t *testing.T, expected string) {
	err := recover()
	if !strings.Contains(fmt.Sprintf("%s", err), expected) {
		t.Fatalf("%s", err)
	}
}
