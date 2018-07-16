package mock

import (
	"strings"
	"testing"
)

type EqualMe struct {
	s string
}

func (x *EqualMe) Equal(y *EqualMe) bool {
	return strings.ToLower(x.s) == strings.ToLower(y.s)
}

type MyMock struct {
	Mock
}

func (m *MyMock) Call(e *EqualMe) {
	m.Called(e)
}

func TestCompareUsingGoCmp(t *testing.T) {
	sut := &MyMock{}

	foo := &EqualMe{"foo"}
	FOO := &EqualMe{"FOO"}

	sut.When("Call", foo).Return(struct{}{})

	sut.Call(FOO)

	sut.Verify()
}
