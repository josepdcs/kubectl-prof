package api

import "github.com/samber/lo"

// ProgrammingLanguage represents a supported programming language for profiling.
type ProgrammingLanguage string

const (
	Java          ProgrammingLanguage = "java"    // Java represents the Java programming language.
	Go            ProgrammingLanguage = "go"      // Go represents the Go programming language.
	Python        ProgrammingLanguage = "python"  // Python represents the Python programming language.
	Ruby          ProgrammingLanguage = "ruby"    // Ruby represents the Ruby programming language.
	Node          ProgrammingLanguage = "node"    // Node represents Node.js (JavaScript runtime).
	Clang         ProgrammingLanguage = "clang"   // Clang represents C language compiled with Clang.
	ClangPlusPlus ProgrammingLanguage = "clang++" // ClangPlusPlus represents C++ language compiled with Clang.
	Rust          ProgrammingLanguage = "rust"    // Rust represents the Rust programming language.
	FakeLang      ProgrammingLanguage = "fake"    // FakeLang represents a fake language for testing purposes.
)

var (
	// supportedLangs contains all supported programming languages for profiling.
	supportedLangs = []ProgrammingLanguage{Java, Go, Python, Ruby, Node, Clang, ClangPlusPlus, Rust}
)

// AvailableLanguages returns the list of all supported programming languages.
func AvailableLanguages() []ProgrammingLanguage {
	return supportedLangs
}

// IsSupportedLanguage checks if the given language string is supported for profiling.
// It returns true if the language is in the list of available languages or is the fake language for testing.
func IsSupportedLanguage(lang string) bool {
	if lang == string(FakeLang) {
		return true
	}
	return lo.Contains(AvailableLanguages(), ProgrammingLanguage(lang))
}
