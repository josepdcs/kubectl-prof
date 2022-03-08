package cli

import "fmt"

type Printer interface {
	Print(str string)
	PrintSuccess()
	PrintError()
}

type dryRunPrinter struct {
	dryRun bool
}

func (p *dryRunPrinter) Print(str string) {
	if !p.dryRun {
		fmt.Print(str)
	}
}

func (p *dryRunPrinter) PrintSuccess() {
	if !p.dryRun {
		fmt.Printf("✔\n")
	}
}

func (p *dryRunPrinter) PrintError() {
	fmt.Printf("❌\n")
}

func NewPrinter(dryRun bool) Printer {
	return &dryRunPrinter{
		dryRun: dryRun,
	}
}
