// Package flamegraph provides basic util methods for working with flame graphs by using the well known tools as flamegraph.pl and more.
package flamegraph

import (
	"bytes"
	"fmt"
	"os"
	"reflect"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
)

// FlameGrapherScript encapsulates the flamegraph.pl script (with its complete path) and the options that can be passed to it.
// See all options with default values in https://github.com/brendangregg/FlameGraph/blob/master/flamegraph.pl
type FlameGrapherScript struct {
	path       string
	title      string
	subtitle   string
	width      string
	height     string
	minWidth   string
	fontType   string
	fontSize   string
	countName  string
	nameType   string
	colors     string
	bgColors   string
	hash       bool
	reverse    bool
	inverted   bool
	flameChart bool
	negate     bool
}

// Option represents an option of the FlameGrapherScript.
type Option func(s *FlameGrapherScript)

// NewFlameGrapherScript returns a new FlameGrapherScript.
// Several Option can be provided in order to override the default ones.
func NewFlameGrapherScript(options ...Option) *FlameGrapherScript {
	g := &FlameGrapherScript{
		path:     "/app/FlameGraph/flamegraph.pl",
		title:    "CPU Flamegraph",
		width:    "1800",
		height:   "16",
		fontType: "Verdana",
		fontSize: "12",
	}

	for _, option := range options {
		option(g)
	}

	return g
}

// WithPath sets the full path
func WithPath(path string) Option {
	return func(s *FlameGrapherScript) {
		if stringUtils.IsNotBlank(path) {
			s.path = path
		}
	}
}

// WithTitle sets the title
func WithTitle(title string) Option {
	return func(s *FlameGrapherScript) {
		if stringUtils.IsNotBlank(title) {
			s.title = title
		}
	}
}

// WithSubtitle sets the subtitle
func WithSubtitle(subtitle string) Option {
	return func(s *FlameGrapherScript) {
		s.subtitle = subtitle
	}
}

// WithWidth sets the width
func WithWidth(width string) Option {
	return func(s *FlameGrapherScript) {
		if stringUtils.IsNotBlank(width) && stringUtils.IsNumeric(width) {
			s.width = width
		}
	}
}

// WithHeight sets the height
func WithHeight(height string) Option {
	return func(s *FlameGrapherScript) {
		if stringUtils.IsNotBlank(height) && stringUtils.IsNumeric(height) {
			s.height = height
		}
	}
}

// WithMinWidth sets the min width
func WithMinWidth(minWidth string) Option {
	return func(s *FlameGrapherScript) {
		if stringUtils.IsNumeric(minWidth) {
			s.minWidth = minWidth
		}
	}
}

// WithFontType sets the font type
func WithFontType(fontType string) Option {
	return func(s *FlameGrapherScript) {
		s.fontType = fontType
	}
}

// WithFontSize sets the font size
func WithFontSize(fontSize string) Option {
	return func(s *FlameGrapherScript) {
		if stringUtils.IsNumeric(fontSize) {
			s.fontSize = fontSize
		}
	}
}

// WithCountName sets the count name
func WithCountName(countName string) Option {
	return func(s *FlameGrapherScript) {
		s.countName = countName
	}
}

// WithNameType sets the name tyoe
func WithNameType(nameType string) Option {
	return func(s *FlameGrapherScript) {
		s.nameType = nameType
	}
}

// WithColors sets the colors
func WithColors(colors string) Option {
	return func(s *FlameGrapherScript) {
		s.colors = colors
	}
}

// WithBgColors sets the bg colors
func WithBgColors(bgColors string) Option {
	return func(s *FlameGrapherScript) {
		s.bgColors = bgColors
	}
}

// WithHash sets hash
func WithHash(hash bool) Option {
	return func(s *FlameGrapherScript) {
		s.hash = hash
	}
}

// WithReverse sets reverse
func WithReverse(reverse bool) Option {
	return func(s *FlameGrapherScript) {
		s.reverse = reverse
	}
}

// WithInverted sets inverted
func WithInverted(inverted bool) Option {
	return func(s *FlameGrapherScript) {
		s.inverted = inverted
	}
}

// WithFlameChart sets flame chart
func WithFlameChart(flameChart bool) Option {
	return func(s *FlameGrapherScript) {
		s.flameChart = flameChart
	}
}

// WithNegate sets negate
func WithNegate(negate bool) Option {
	return func(s *FlameGrapherScript) {
		s.negate = negate
	}
}

// StackSamplesToFlameGraph converts input file, which contains stack samples,
// to flame graph output file, by using Brendan Gregg's tool flamegraph.pl
func (g *FlameGrapherScript) StackSamplesToFlameGraph(inputFileName string, outputFileName string) error {
	inputFile, err := os.Open(inputFileName)
	if err != nil {
		return err
	}

	defer func(inputFile *os.File) {
		err := inputFile.Close()
		if err != nil {
			log.ErrorLogLn(fmt.Sprintf("error closing input file: %s", err))
			return
		}
	}(inputFile)

	outputFile, err := os.Create(outputFileName)
	if err != nil {
		return err
	}

	defer func(outputFile *os.File) {
		err := outputFile.Close()
		if err != nil {
			log.ErrorLogLn(fmt.Sprintf("error closing output file: %s", err))
			return
		}
	}(outputFile)

	var stderr bytes.Buffer
	cmd := exec.Command(g.path, scriptToArguments(g)...)
	cmd.Stdin = inputFile
	cmd.Stdout = outputFile
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.ErrorLogLn(stderr.String())
	}

	return err
}

// scriptToArguments returns the arguments to be passed to the flamegraph.pl
func scriptToArguments(g *FlameGrapherScript) []string {
	args := make([]string, 0, reflect.TypeOf(*g).NumField()*2)

	args = append(args,
		"--title", g.title,
		"--width", g.width,
		"--height", g.height,
		"--fonttype", g.fontType,
		"--fontsize", g.fontSize,
	)

	if stringUtils.IsNotBlank(g.subtitle) {
		args = append(args, "--subtitle", g.subtitle)
	}
	if stringUtils.IsNotBlank(g.minWidth) {
		args = append(args, "--minwidth", g.minWidth)
	}
	if stringUtils.IsNotBlank(g.countName) {
		args = append(args, "--countname", g.countName)
	}
	if stringUtils.IsNotBlank(g.nameType) {
		args = append(args, "--nametype", g.nameType)
	}
	if stringUtils.IsNotBlank(g.colors) {
		args = append(args, "--colors", g.colors)
	}
	if stringUtils.IsNotBlank(g.bgColors) {
		args = append(args, "--bgcolors", g.bgColors)
	}
	if g.hash {
		args = append(args, "--hash")
	}
	if g.reverse {
		args = append(args, "--reverse")
	}
	if g.inverted {
		args = append(args, "--inverted")
	}
	if g.flameChart {
		args = append(args, "--flamechart")
	}
	if g.negate {
		args = append(args, "--negate")
	}

	return args
}
