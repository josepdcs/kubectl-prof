package api

import "github.com/samber/lo"

type ProgrammingLanguage string

const (
	Java          ProgrammingLanguage = "java"
	Go            ProgrammingLanguage = "go"
	Python        ProgrammingLanguage = "python"
	Ruby          ProgrammingLanguage = "ruby"
	Node          ProgrammingLanguage = "node"
	Clang         ProgrammingLanguage = "clang"
	ClangPlusPlus ProgrammingLanguage = "clang++"
	Rust          ProgrammingLanguage = "rust"
	Php           ProgrammingLanguage = "php"
	FakeLang      ProgrammingLanguage = "fake"
)

var (
	supportedLangs = []ProgrammingLanguage{Java, Go, Python, Ruby, Node, Clang, ClangPlusPlus, Rust, Php}
)

func AvailableLanguages() []ProgrammingLanguage {
	return supportedLangs
}

func IsSupportedLanguage(lang string) bool {
	if lang == string(FakeLang) {
		return true
	}
	return lo.Contains(AvailableLanguages(), ProgrammingLanguage(lang))
}
