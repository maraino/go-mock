package mock

import (
	"strings"
	"testing"
)

type EqualMe struct {
	s string
}

type EqualMeToo struct {
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

func (m *MyMock) CallToo(e *EqualMeToo) {
	m.Called(e)
}

func TestCompareUsingGoCmp(t *testing.T) {
	sut := &MyMock{}

	foo := &EqualMe{"foo"}
	FOO := &EqualMe{"FOO"}

	sut.When("Call", foo).Return(struct{}{}).Times(1)

	sut.Call(FOO)

	ok, err := sut.Verify()
	if !ok {
		t.Error(err)
	}

}

