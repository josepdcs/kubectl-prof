package job

import (
	"errors"

	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"

	"github.com/josepdcs/kubectl-prof/api"
)

const (
	command       = "/app/agent"
	baseImageName = "josepdcs/kubectl-prof"
	ContainerName = "kubectl-prof"
	LabelID       = "kubectl-prof/id"
)

// Creator defines the method for creating the profiling job according the programming language.
type Creator interface {
	Create(targetPod *apiv1.Pod, cfg *config.ProfilerConfig) (string, *batchv1.Job, error)
}

// jobCreatorLink defines the interface for the job creator chain of responsibility.
type jobCreatorLink interface {
	// next sets the next link in the chain.
	next(link jobCreatorLink) jobCreatorLink
	// get returns the Creator implementation according the programming language and profiling tool.
	get(lang api.ProgrammingLanguage, tool api.ProfilingTool) (Creator, error)
}

// baseJobCreatorLink is the base implementation for the job creator chain of responsibility.
type baseJobCreatorLink struct {
	nextLink jobCreatorLink
}

// next sets the next link in the chain.
func (b *baseJobCreatorLink) next(link jobCreatorLink) jobCreatorLink {
	b.nextLink = link
	return link
}

// get returns the Creator implementation according the programming language and profiling tool from the next link in the chain.
func (b *baseJobCreatorLink) get(lang api.ProgrammingLanguage, tool api.ProfilingTool) (Creator, error) {
	if b.nextLink != nil {
		return b.nextLink.get(lang, tool)
	}
	return nil, errors.New("got language without job creator")
}

// jvmCreatorLink is the job creator link for Java.
type jvmCreatorLink struct {
	baseJobCreatorLink
}

// get returns the Creator implementation for Java.
func (l *jvmCreatorLink) get(lang api.ProgrammingLanguage, tool api.ProfilingTool) (Creator, error) {
	if lang == api.Java {
		return &jvmCreator{}, nil
	}
	return l.baseJobCreatorLink.get(lang, tool)
}

// nativeCreatorLink is the job creator link for Go, Clang, ClangPlusPlus and Node.
type nativeCreatorLink struct {
	baseJobCreatorLink
}

// get returns the Creator implementation for Go, Clang, ClangPlusPlus and Node.
func (l *nativeCreatorLink) get(lang api.ProgrammingLanguage, tool api.ProfilingTool) (Creator, error) {
	if lang == api.Go || lang == api.Clang || lang == api.ClangPlusPlus || lang == api.Node {
		if tool == api.Perf {
			return &perfCreator{}, nil
		}
		if tool == api.NodeDummy {
			return &dummyCreator{}, nil
		}
		if tool == api.Btf {
			return &btfCreator{}, nil
		}
		if tool == api.GoPprof {
			return &pprofCreator{}, nil
		}
		return &bpfCreator{}, nil
	}
	return l.baseJobCreatorLink.get(lang, tool)
}

// rustCreatorLink is the job creator link for Rust.
type rustCreatorLink struct {
	baseJobCreatorLink
}

// get returns the Creator implementation for Rust.
func (l *rustCreatorLink) get(lang api.ProgrammingLanguage, tool api.ProfilingTool) (Creator, error) {
	if lang == api.Rust {
		if tool == api.CargoFlame {
			return &rustCreator{}, nil
		}
		if tool == api.Perf {
			return &perfCreator{}, nil
		}
		if tool == api.Btf {
			return &btfCreator{}, nil
		}
		return &bpfCreator{}, nil
	}
	return l.baseJobCreatorLink.get(lang, tool)
}

// pythonCreatorLink is the job creator link for Python.
type pythonCreatorLink struct {
	baseJobCreatorLink
}

// get returns the Creator implementation for Python.
func (l *pythonCreatorLink) get(lang api.ProgrammingLanguage, tool api.ProfilingTool) (Creator, error) {
	if lang == api.Python {
		return &pythonCreator{}, nil
	}
	return l.baseJobCreatorLink.get(lang, tool)
}

// rubyCreatorLink is the job creator link for Ruby.
type rubyCreatorLink struct {
	baseJobCreatorLink
}

// get returns the Creator implementation for Ruby.
func (l *rubyCreatorLink) get(lang api.ProgrammingLanguage, tool api.ProfilingTool) (Creator, error) {
	if lang == api.Ruby {
		return &rubyCreator{}, nil
	}
	return l.baseJobCreatorLink.get(lang, tool)
}

// phpCreatorLink is the job creator link for PHP.
type phpCreatorLink struct {
	baseJobCreatorLink
}

// get returns the Creator implementation for PHP.
func (l *phpCreatorLink) get(lang api.ProgrammingLanguage, tool api.ProfilingTool) (Creator, error) {
	if lang == api.PHP {
		return &phpCreator{}, nil
	}
	return l.baseJobCreatorLink.get(lang, tool)
}

// dotnetCreatorLink is the job creator link for DotNet.
type dotnetCreatorLink struct {
	baseJobCreatorLink
}

// get returns the Creator implementation for DotNet.
func (l *dotnetCreatorLink) get(lang api.ProgrammingLanguage, tool api.ProfilingTool) (Creator, error) {
	if lang == api.DotNet {
		return &dotnetCreator{}, nil
	}
	return l.baseJobCreatorLink.get(lang, tool)
}

// fakeCreatorLink is the job creator link for FakeLang.
type fakeCreatorLink struct {
	baseJobCreatorLink
}

// get returns the Creator implementation for FakeLang.
func (l *fakeCreatorLink) get(lang api.ProgrammingLanguage, tool api.ProfilingTool) (Creator, error) {
	if lang == api.FakeLang {
		return &fakeCreator{}, nil
	}
	return l.baseJobCreatorLink.get(lang, tool)
}

// NewCreator returns the Creator implementation according the programming language.
func NewCreator(lang api.ProgrammingLanguage, tool api.ProfilingTool) (Creator, error) {
	chain := &jvmCreatorLink{}
	chain.next(&nativeCreatorLink{}).
		next(&rustCreatorLink{}).
		next(&pythonCreatorLink{}).
		next(&rubyCreatorLink{}).
		next(&phpCreatorLink{}).
		next(&dotnetCreatorLink{}).
		next(&fakeCreatorLink{})

	return chain.get(lang, tool)
}
