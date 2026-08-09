package forms

import "charm.land/huh/v2"

func LoadCliFrameworkForm(cliFramework, selectedProject *string) *huh.Group {
	form := huh.NewGroup(
		huh.NewSelect[string]().
			Title("What kind of CLI application?\n").
			Options(
				huh.NewOption("Basic Input Form (Huh)", "cli-huh"),
				huh.NewOption("Interactive TUI (Bubble Tea)", "cli-bubbletea"),
				huh.NewOption("Standard Flags Only (Cobra)", "cli-cobra"),
			).
			Value(cliFramework),
	).WithHideFunc(func() bool {
		return *selectedProject != "cli_project"
	})
	return form
}
