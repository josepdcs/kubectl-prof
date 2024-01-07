package exec

import (
	"os/exec"
)

type FakeCommander interface {
	On(methodName string) *FakeMethod

	Commander
}

type Fake struct {
	fakeMethods map[string]*FakeMethod
}

type FakeMethod struct {
	*fakeMethod
}

type fakeMethod struct {
	fakeReturnValues []interface{}
	invokes          int
	indexCommand     int
}

func NewFakeCommander() FakeCommander {
	return &Fake{
		fakeMethods: make(map[string]*FakeMethod),
	}
}

func (f *Fake) On(methodName string) *FakeMethod {
	if _, ok := f.fakeMethods[methodName]; !ok {
		f.fakeMethods[methodName] = &FakeMethod{&fakeMethod{}}
	}
	if f.fakeMethods[methodName].fakeReturnValues == nil {
		f.fakeMethods[methodName].fakeReturnValues = make([]interface{}, 0)
	}
	return f.fakeMethods[methodName]
}

func (f *FakeMethod) Return(fakeReturnValues ...interface{}) *FakeMethod {
	f.fakeReturnValues = append(f.fakeReturnValues, fakeReturnValues)
	return f
}

func (f *FakeMethod) InvokedTimes() int {
	return f.invokes
}

func (f *Fake) Command(string, ...string) *exec.Cmd {
	f.fakeMethods["Command"].invokes++
	if f.fakeMethods["Command"].fakeReturnValues != nil && len(f.fakeMethods["Command"].fakeReturnValues) > 0 {
		f.fakeMethods["Command"].indexCommand++
		return f.fakeMethods["Command"].fakeReturnValues[f.fakeMethods["Command"].indexCommand-1].([]interface{})[0].(*exec.Cmd)
	}
	return nil
}

func (f *Fake) Execute(*exec.Cmd) (int, []byte, error) {
	f.fakeMethods["Execute"].invokes++
	if f.fakeMethods["Execute"].fakeReturnValues != nil && len(f.fakeMethods["Execute"].fakeReturnValues) > 0 {
		f.fakeMethods["Execute"].indexCommand++
		return f.fakeMethods["Execute"].fakeReturnValues[f.fakeMethods["Execute"].indexCommand-1].([]interface{})[0].(int),
			f.fakeMethods["Execute"].fakeReturnValues[f.fakeMethods["Execute"].indexCommand-1].([]interface{})[1].([]byte),
			f.fakeMethods["Execute"].fakeReturnValues[f.fakeMethods["Execute"].indexCommand-1].([]interface{})[2].(error)
	}
	return 0, nil, nil
}
