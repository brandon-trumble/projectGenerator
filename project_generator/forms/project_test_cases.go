package forms

import "charm.land/huh/v2"

func LoadTestCasesForm(allowTestCases *bool, selectedProject *string) *huh.Group {
	form := huh.NewGroup(
		huh.NewConfirm().
			Title("Include test cases?").
			Affirmative("Yes").
			Negative("No").
			Value(allowTestCases),
	).WithHideFunc(func() bool {
		return *selectedProject != "http_backend"
	})

	return form
}
