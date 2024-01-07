package exec

import "os/exec"

type FakeCommander interface {
	Return(fakeReturnValues ...interface{}) *Fake
	On(methodName string) *Fake
	InvokedTimes(methodName string) int

	//ReturnCmd(fakeCommand *exec.Cmd) *Fake

	Commander
}

type Fake struct {
	*fakeCommander
}

type fakeCommander struct {
	fakeReturnValues [][]interface{}
	methodName       string
	invokes          map[string]int

	//fakeCommand  []*exec.Cmd
	indexCommand map[string]int
}

func NewFakeCommander() FakeCommander {
	return &Fake{
		fakeCommander: &fakeCommander{
			fakeReturnValues: make([][]interface{}, 0),
			invokes:          make(map[string]int),
			//fakeCommand:      make([]*exec.Cmd, 0),
			indexCommand: make(map[string]int),
		},
	}
}

func (f *Fake) Return(fakeReturnValues ...interface{}) *Fake {
	f.fakeReturnValues = append(f.fakeReturnValues, fakeReturnValues)
	return f
}

func (f *Fake) On(methodName string) *Fake {
	f.methodName = methodName
	return f
}

func (f *Fake) InvokedTimes(methodName string) int {
	return f.invokes[methodName]
}

/*func (f *Fake) ReturnCmd(fakeCommand *exec.Cmd) *Fake {
	f.fakeCommand = append(f.fakeCommand, fakeCommand)
	return f
}*/

func (f *Fake) Command(string, ...string) *exec.Cmd {
	f.invokes["Command"]++
	if f.methodName == "Command" && f.fakeReturnValues != nil && len(f.fakeReturnValues) > 0 {
		f.indexCommand[f.methodName]++
		/*if f.indexCommand[f.methodName] >= len(f.fakeReturnValues[f.indexCommand[f.methodName]]) {
			return nil
		}*/
		return f.fakeReturnValues[f.indexCommand[f.methodName]-1][0].(*exec.Cmd)
	}
	return nil
}

func (f *Fake) Execute(*exec.Cmd) (int, []byte, error) {
	f.invokes["Execute"]++
	if f.methodName == "Execute" && f.fakeReturnValues != nil && len(f.fakeReturnValues) > 0 {
		if f.indexCommand[f.methodName] >= len(f.fakeReturnValues[f.indexCommand[f.methodName]]) {
			return 0, nil, nil
		}
		return f.fakeReturnValues[f.indexCommand[f.methodName]][0].(int),
			f.fakeReturnValues[f.indexCommand[f.methodName]][1].([]byte),
			f.fakeReturnValues[f.indexCommand[f.methodName]][2].(error)
	}
	f.indexCommand[f.methodName]++
	return 0, nil, nil
}
