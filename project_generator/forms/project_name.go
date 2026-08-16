package forms

import (
	"errors"

	"projectGenerator/project_generator"

	"charm.land/huh/v2"
)

func LoadProjectNameForm(projectName *string, allowProjectName *bool) *huh.Group {
	form := huh.NewGroup(
		huh.NewInput().
			Title("Enter project name\n").
			// the name becomes a directory and a module path, so it has to be
			// checked here rather than failing later inside "go mod init"
			Validate(project_generator.ValidateProjectName).
			Placeholder("EX: funApi").
			Value(projectName),
		huh.NewConfirm().
			Title("Confirm project name?").
			Affirmative("Yes").
			Negative("No").
			Value(allowProjectName).
			Validate(func(b bool) error {
				if !*allowProjectName {
					return errors.New("please enter a new name. Press shift+tab")
				}
				return nil
			}),
	)
	return form
}
