package forms

import (
	"errors"

	"charm.land/huh/v2"
)

func LoadProjectTypeForm(selectedProject *string, allowProjectType *bool) *huh.Group {
	form := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Choose a project\n").
			Options(
				huh.NewOption("Http-Backend", "http_backend"),
				huh.NewOption("Cli-Application", "cli_project"),
				huh.NewOption("Empty-Project", "empty_project"),
			).Value(selectedProject),

		huh.NewConfirm().
			Title("Confirm project type?").
			Affirmative("Yes").
			Negative("No").
			Value(allowProjectType).
			Validate(func(b bool) error {
				if !*allowProjectType {
					return errors.New("please select a new project type. Please press shift+tab")
				}
				return nil
			}),
	)
	return form
}
