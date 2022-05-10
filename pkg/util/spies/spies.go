// Package spies provides test spies for go test, in the vien of testify's mocks.
// Roughly speaking, define a spy like this:
//  type MySpy {
//  	*spies.Spy
//  }
//
//  func NewMySpy() *MySpy {
//    return &MySpy{ NewSpy() }
//  }
//
//  func (my *MySpy) InterfaceMethod(with, some, args int) (ret string, err error) {
//    res := my.Called(with, some, args)
//    return res.String(0), res.Error(1)
//  }
// Use your spy in tests like:
//   my.MatchMethod("InterfaceMethod", spies.AnyArgs, "called", nil)
// ... which will return "called" with a nil error whenever InterfaceMethod is called.
// Several calls to Match and MatchMethod can be made in a row - the first match wins.
// Then you can check calls by calling my.Calls to my.CallsTo
package spies

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/stretchr/testify/mock"
)

type (
	matcher struct {
		pred   func(string, mock.Arguments) bool
		result mock.Arguments
	}

	Call struct {
		method string
		args   mock.Arguments
		res    results
	}

	// A Spy is a type for use in testing - it's intended to be embedded in spy
	// implementations.
	Spy struct {
		matchers []matcher
		calls    []Call
		sync.RWMutex
	}

	// A PassedArger implements PassedArgs
	PassedArger interface {
		PassedArgs() mock.Arguments
	}
)

// NewSpy makes a Spy
func NewSpy() *Spy {
	return &Spy{
		matchers: []matcher{},
		calls:    []Call{},
	}
}

// Always is an always-true predicate
func Always(string, mock.Arguments) bool {
	return true
}

// AnyArgs is an always-true predicate for MethodMatch
func AnyArgs(mock.Arguments) bool {
	return true
}

// Once is a convenience for CallCount(1)
func Once() func(mock.Arguments) bool {
	return CallCount(1)
}

// CallCount constructs a predicate that allows a method to be called a certain number of times.
func CallCount(n int) func(mock.Arguments) bool {
	return func(mock.Arguments) bool {
		n--
		return n >= 0
	}
}

func (s *Spy) String() string {
	str := "Calls: "
	s.RLock()
	defer s.RUnlock()
	for _, c := range s.calls {
		str += c.String() + "\n"
	}
	return str
}

func (c Call) String() string {
	return fmt.Sprintf("%s(%s) -> (%s)", c.method, c.args, c.res)
}

func (c Call) PassedArgs() mock.Arguments {
	as := make(mock.Arguments, len(c.args))
	copy(as, c.args)
	return as
}

// Match records an arbitrary predicate to match against a method Call
func (s *Spy) Match(pred func(string, mock.Arguments) bool, result ...interface{}) {
	s.matchers = append(s.matchers, matcher{pred: pred, result: mock.Arguments(result)})
}

// MatchMethod records a predicate limited to a specific method name
func (s *Spy) MatchMethod(method string, pred func(mock.Arguments) bool, result ...interface{}) {
	s.matchers = append(s.matchers, matcher{
		pred: func(m string, as mock.Arguments) bool {
			if m != method {
				return false
			}
			return pred(as)
		},
		result: mock.Arguments(result),
	})
}

// Any records that any Call to method get result as a reply
func (s *Spy) Any(method string, result ...interface{}) {
	s.matchers = append(s.matchers,
		matcher{
			pred: func(m string, a mock.Arguments) bool {
				return method == m
			},
			result: mock.Arguments(result),
		})
}

func (s *Spy) findArgs(functionName string, args mock.Arguments) results {
	for _, m := range s.matchers {
		if m.pred(functionName, args) {
			if m.result == nil {
				return presentResult{mock.Arguments{}}
			}
			return presentResult{m.result}
		}
	}
	return missingResult{}
}

// Called is used by embedders of Spy to indicate that the method is called.
func (s *Spy) Called(argList ...interface{}) results {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("Couldn't get caller info")
	}

	functionPath := runtime.FuncForPC(pc).Name()
	//Next four lines are required to use GCCGO function naming conventions.
	//For Ex:  github_com_docker_libkv_store_mock.WatchTree.pN39_github_com_docker_libkv_store_mock.Mock
	//uses interface information unlike golang github.com/docker/libkv/store/mock.(*Mock).WatchTree
	//With GCCGO we need to remove interface information starting from pN<dd>.
	re := regexp.MustCompile("\\.pN\\d+_")
	if re.MatchString(functionPath) {
		functionPath = re.Split(functionPath, -1)[0]
	}
	parts := strings.Split(functionPath, ".")
	functionName := parts[len(parts)-1]

	args := mock.Arguments(argList)

	found := s.findArgs(functionName, args)

	s.Lock()
	defer s.Unlock()
	s.calls = append(s.calls, Call{functionName, args, found})
	return found
}

// CallsMatching returns a list of calls for with f() returns true.
func (s *Spy) CallsMatching(f func(name string, args mock.Arguments) bool) []Call {
	s.RLock()
	defer s.RUnlock()
	calls := []Call{}
	for _, c := range s.calls {
		if f(c.method, c.args) {
			calls = append(calls, c)
		}
	}
	return calls
}

// CallsTo returns the calls to the named method
func (s *Spy) CallsTo(name string) []Call {
	return s.CallsMatching(func(n string, _ mock.Arguments) bool {
		return name == n
	})
}

// Calls returns all the calls made to the spy
func (s *Spy) Calls() []Call {
	s.RLock()
	s.RUnlock()
	cs := make([]Call, len(s.calls))
	copy(cs, s.calls)
	return cs
}
