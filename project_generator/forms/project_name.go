package forms

import (
	"errors"

	"charm.land/huh/v2"
)

func LoadProjectNameForm(projectName *string, allowProjectName *bool) *huh.Group {
	form := huh.NewGroup(
		huh.NewInput().
			Title("Enter project name\n").
			Validate(func(s string) error {
				if s == "" {
					return errors.New("project name is invalid. Please enter a new name")
				}
				return nil
			}).
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
