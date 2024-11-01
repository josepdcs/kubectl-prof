package cli

import "fmt"

// Printer defines the methods for printing messages
type Printer interface {
	Print(str string)
	PrintSuccess()
	PrintError()
}

// NewPrinter returns new instance of Printer
func NewPrinter(dryRun bool) Printer {
	return &dryRunPrinter{
		dryRun: dryRun,
	}
}

// NewPrinterWithTargetPod returns new instance of Printer with target pod
func NewPrinterWithTargetPod(dryRun bool, targetPod string) Printer {
	return &dryRunPrinter{
		dryRun:    dryRun,
		targetPod: targetPod,
	}
}

type dryRunPrinter struct {
	dryRun    bool
	targetPod string
}

func (p *dryRunPrinter) Print(str string) {
	if !p.dryRun {
		if p.targetPod != "" {
			str = fmt.Sprintf("[%s] %s", p.targetPod, str)
		}
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
