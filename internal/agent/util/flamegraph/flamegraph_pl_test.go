package flamegraph

import (
	"github.com/josepdcs/kubectl-prof/internal/agent/testdata"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestNewGrapherScript(t *testing.T) {
	type args struct {
		options []Option
	}
	tests := []struct {
		name  string
		given func() args
		when  func(args) *FlameGrapherScript
		then  func(t *testing.T, g *FlameGrapherScript)
	}{
		{
			name: "should new default FlameGrapherScript",
			given: func() args {
				return args{}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "16",
					fontType: "Verdana",
					fontSize: "12",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with path",
			given: func() args {
				return args{
					options: []Option{WithPath("/other/path")},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/other/path",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "16",
					fontType: "Verdana",
					fontSize: "12",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with path ignored when empty",
			given: func() args {
				return args{
					options: []Option{WithPath("")},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "16",
					fontType: "Verdana",
					fontSize: "12",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with title",
			given: func() args {
				return args{
					options: []Option{WithTitle("Title")},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "Title",
					width:    "1800",
					height:   "16",
					fontType: "Verdana",
					fontSize: "12",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with title ignored when empty",
			given: func() args {
				return args{
					options: []Option{WithTitle("")},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "16",
					fontType: "Verdana",
					fontSize: "12",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with subtitle",
			given: func() args {
				return args{
					options: []Option{WithSubtitle("Subtitle")},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					subtitle: "Subtitle",
					width:    "1800",
					height:   "16",
					fontType: "Verdana",
					fontSize: "12",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with width",
			given: func() args {
				return args{
					options: []Option{WithWidth("1000")},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1000",
					height:   "16",
					fontType: "Verdana",
					fontSize: "12",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with width ignored when empty",
			given: func() args {
				return args{
					options: []Option{WithWidth("")},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "16",
					fontType: "Verdana",
					fontSize: "12",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with width ignored when not numeric",
			given: func() args {
				return args{
					options: []Option{WithWidth("no-numeric")},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "16",
					fontType: "Verdana",
					fontSize: "12",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with height",
			given: func() args {
				return args{
					options: []Option{WithHeight("20")},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "20",
					fontType: "Verdana",
					fontSize: "12",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with height ignored when empty",
			given: func() args {
				return args{
					options: []Option{WithHeight("")},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "16",
					fontType: "Verdana",
					fontSize: "12",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with height ignored when not numeric",
			given: func() args {
				return args{
					options: []Option{WithHeight("no-numeric")},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "16",
					fontType: "Verdana",
					fontSize: "12",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with minWidth",
			given: func() args {
				return args{
					options: []Option{WithMinWidth("100")},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "16",
					minWidth: "100",
					fontType: "Verdana",
					fontSize: "12",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with minWidth ignored when not numeric",
			given: func() args {
				return args{
					options: []Option{WithMinWidth("no-numeric")},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "16",
					fontType: "Verdana",
					fontSize: "12",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with fontType",
			given: func() args {
				return args{
					options: []Option{WithFontType("FontType")},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "16",
					fontType: "FontType",
					fontSize: "12",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with fontSize",
			given: func() args {
				return args{
					options: []Option{WithFontSize("10")},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "16",
					fontType: "Verdana",
					fontSize: "10",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with fontSize ignored when not numeric",
			given: func() args {
				return args{
					options: []Option{WithFontSize("no-numeric")},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "16",
					fontType: "Verdana",
					fontSize: "12",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with countName",
			given: func() args {
				return args{
					options: []Option{WithCountName("CountName")},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:      "/app/FlameGraph/flamegraph.pl",
					title:     "CPU Flamegraph",
					width:     "1800",
					height:    "16",
					fontType:  "Verdana",
					fontSize:  "12",
					countName: "CountName",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with nameType",
			given: func() args {
				return args{
					options: []Option{WithNameType("NameType")},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "16",
					fontType: "Verdana",
					fontSize: "12",
					nameType: "NameType",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with colors",
			given: func() args {
				return args{
					options: []Option{WithColors("Colors")},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "16",
					fontType: "Verdana",
					fontSize: "12",
					colors:   "Colors",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with bgColors",
			given: func() args {
				return args{
					options: []Option{WithBgColors("BgColors")},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "16",
					fontType: "Verdana",
					fontSize: "12",
					bgColors: "BgColors",
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with hash",
			given: func() args {
				return args{
					options: []Option{WithHash(true)},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "16",
					fontType: "Verdana",
					fontSize: "12",
					hash:     true,
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with reverse",
			given: func() args {
				return args{
					options: []Option{WithReverse(true)},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "16",
					fontType: "Verdana",
					fontSize: "12",
					reverse:  true,
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with inverted",
			given: func() args {
				return args{
					options: []Option{WithInverted(true)},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "16",
					fontType: "Verdana",
					fontSize: "12",
					inverted: true,
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with flameChart",
			given: func() args {
				return args{
					options: []Option{WithFlameChart(true)},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:       "/app/FlameGraph/flamegraph.pl",
					title:      "CPU Flamegraph",
					width:      "1800",
					height:     "16",
					fontType:   "Verdana",
					fontSize:   "12",
					flameChart: true,
				}
				assert.Equal(t, expected, g)
			},
		},
		{
			name: "should new FlameGrapherScript with negate",
			given: func() args {
				return args{
					options: []Option{WithNegate(true)},
				}
			},
			when: func(args args) *FlameGrapherScript {
				return NewFlameGrapherScript(args.options...)
			},
			then: func(t *testing.T, g *FlameGrapherScript) {
				expected := &FlameGrapherScript{
					path:     "/app/FlameGraph/flamegraph.pl",
					title:    "CPU Flamegraph",
					width:    "1800",
					height:   "16",
					fontType: "Verdana",
					fontSize: "12",
					negate:   true,
				}
				assert.Equal(t, expected, g)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			args := tt.given()

			// When
			result := tt.when(args)

			// Then
			tt.then(t, result)
		})
	}
}

func TestGrapherScript_StackSamplesToFlameGraph(t *testing.T) {
	type fields struct {
		g *FlameGrapherScript
	}
	type args struct {
		rawFileName        string
		flameGraphFileName string
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) error
		then  func(t *testing.T, err error, f fields)
		after func()
	}{
		{
			name: "should fail when input file not exists",
			given: func() (fields, args) {
				return fields{
						g: NewFlameGrapherScript(),
					}, args{
						rawFileName:        "unknown",
						flameGraphFileName: "",
					}
			},
			when: func(fields fields, args args) error {
				return fields.g.StackSamplesToFlameGraph(args.rawFileName, args.flameGraphFileName)
			},
			then: func(t *testing.T, err error, fields fields) {
				require.Error(t, err)
			},
		},
		{
			name: "should fail when output file cannot be created",
			given: func() (fields, args) {
				return fields{
						g: NewFlameGrapherScript(),
					}, args{
						rawFileName:        filepath.Join(testdata.ResultTestDataDir(), "raw.txt"),
						flameGraphFileName: "",
					}
			},
			when: func(fields fields, args args) error {
				return fields.g.StackSamplesToFlameGraph(args.rawFileName, args.flameGraphFileName)
			},
			then: func(t *testing.T, err error, fields fields) {
				require.Error(t, err)
			},
		},
		{
			name: "should fail when script cannot be invoked",
			given: func() (fields, args) {
				return fields{
						g: NewFlameGrapherScript(),
					}, args{
						rawFileName:        filepath.Join(testdata.ResultTestDataDir(), "raw.txt"),
						flameGraphFileName: filepath.Join("/tmp", "flamegrapher_pl_output.svg"),
					}
			},
			when: func(fields fields, args args) error {
				return fields.g.StackSamplesToFlameGraph(args.rawFileName, args.flameGraphFileName)
			},
			then: func(t *testing.T, err error, fields fields) {
				require.Error(t, err)
			},
			after: func() {
				_ = file.Remove(filepath.Join("/tmp", "flamegrapher_pl_output.svg"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			fields, args := tt.given()

			// When
			err := tt.when(fields, args)

			// Then
			tt.then(t, err, fields)

			if tt.after != nil {
				tt.after()
			}
		})
	}
}

func Test_scriptToArguments(t *testing.T) {
	type args struct {
		g *FlameGrapherScript
	}
	tests := []struct {
		name  string
		given func() args
		when  func(args) []string
		then  func(t *testing.T, s []string)

		args args
		want []string
	}{
		{
			name: "With default arguments",
			given: func() args {
				return args{
					NewFlameGrapherScript(),
				}
			},
			when: func(args args) []string {
				return scriptToArguments(args.g)
			},
			then: func(t *testing.T, s []string) {
				expected := []string{
					"--title", "CPU Flamegraph",
					"--width", "1800",
					"--height", "16",
					"--fonttype", "Verdana",
					"--fontsize", "12",
				}
				assert.Equal(t, expected, s)
			},
		},
		{
			name: "With full arguments",
			given: func() args {
				return args{
					NewFlameGrapherScript(
						WithTitle("Title"),
						WithSubtitle("Subtitle"),
						WithWidth("1000"),
						WithHeight("20"),
						WithMinWidth("1"),
						WithFontType("FontType"),
						WithFontSize("20"),
						WithCountName("CountName"),
						WithNameType("NameType"),
						WithColors("Colors"),
						WithBgColors("BgColors"),
						WithHash(true),
						WithReverse(true),
						WithInverted(true),
						WithFlameChart(true),
						WithNegate(true),
					),
				}
			},
			when: func(args args) []string {
				return scriptToArguments(args.g)
			},
			then: func(t *testing.T, s []string) {
				expected := []string{
					"--title", "Title",
					"--subtitle", "Subtitle",
					"--width", "1000",
					"--height", "20",
					"--minwidth", "1",
					"--fonttype", "FontType",
					"--fontsize", "20",
					"--countname", "CountName",
					"--nametype", "NameType",
					"--colors", "Colors",
					"--bgcolors", "BgColors",
					"--hash",
					"--reverse",
					"--inverted",
					"--flamechart",
					"--negate",
				}
				assert.ElementsMatch(t, expected, s)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			args := tt.given()

			// When
			result := tt.when(args)

			// Then
			tt.then(t, result)
		})
	}
}
