package forms

import (
	"errors"

	"charm.land/huh/v2"
)

func LoadProjectDatabaseForm(selectedDatabase, selectedProject *string, allowDatabaseFramework *bool) *huh.Group {
	form := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Choose a database framework\n").
			Options(
				huh.NewOption("PostgreSQL(PGX driver)", "postgres"),
				huh.NewOption("Mysql", "mysql"),
				huh.NewOption("No database", "none"),
			).Value(selectedDatabase),
		huh.NewConfirm().
			Title("Confirm database?").
			Affirmative("Yes").
			Negative("No").
			Value(allowDatabaseFramework).
			Validate(func(b bool) error {
				if !*allowDatabaseFramework {
					return errors.New("please select a new database framework. Please press shift+tab")
				}
				return nil
			}),
	).WithHideFunc(func() bool {
		return *selectedProject != "http_backend"
	})
	return form
}
