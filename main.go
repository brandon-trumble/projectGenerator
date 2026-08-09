package main

import (
	"errors"
	"fmt"
	"projectGenerator/project_generator"
	"projectGenerator/project_generator/forms"
	"runtime"
	"time"

	"charm.land/huh/v2"
	"charm.land/huh/v2/spinner"
	"charm.land/lipgloss/v2"
)

func main() {

	var projectName string
	var selectedProject string
	var selectedDatabase string
	var cliFramework string
	var allowProjectType bool
	var allowDatabaseFramework bool
	var allowProjectName bool
	var allowTestCases bool

	asciColor := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	cancelledText := lipgloss.NewStyle().Foreground(lipgloss.Red)
	successText := lipgloss.NewStyle().Foreground(lipgloss.BrightGreen)
	infoText := lipgloss.NewStyle().Foreground(lipgloss.Color("85"))

	userOs := runtime.GOOS
	fmt.Println(infoText.Render("\ncurrent operating system:", userOs))

	banner := `
    ____             _           __          ______
   / __ \_________  (_)__  _____/ /_        / ____/___  ____
  / /_/ / ___/ __ \/ / _ \/ ___/ __/_______/ / __/ __ \/ __ \
 / ____/ /  / /_/ / /  __/ /__/ /_/_______/ /_/ /  __/ / / /
/_/   /_/   \____/ /\___/\___/\__/        \____/\___/_/ /_/
              /___/
`

	fmt.Println("\n", asciColor.Render(banner))
	// basic form
	form := huh.NewForm(

		forms.LoadProjectNameForm(&projectName, &allowProjectName),

		forms.LoadProjectTypeForm(&selectedProject, &allowProjectType),

		// this will be hidden if they don't choose the http_backend
		forms.LoadProjectDatabaseForm(&selectedDatabase, &selectedProject, &allowDatabaseFramework),

		// this will only show with http_backend selected
		forms.LoadTestCasesForm(&allowTestCases, &selectedProject),

		// this will only show with the command line selected
		forms.LoadCliFrameworkForm(&cliFramework, &selectedProject),
	)

	if err := form.Run(); err != nil {
		// catch user cancellations and print a clean exit message
		if errors.Is(err, huh.ErrUserAborted) {
			fmt.Println(cancelledText.Render("Scaffold cancelled."))
			return
		}

		// catch terminal or rendering errors
		fmt.Println("Error:", err)
		return
	}

	var generationErr error
	var projectDir string
	spinnerErr := spinner.New().
		Type(spinner.Dots).
		Title(" Generating project...").
		Action(func() {
			// NOTE: the time.Sleep is to prevent the huh framework spinner from bugging out due to it not cycling.
			// Without that line of code this is appended to the stdout: ^[]11;rgb:1919/1a1a/1c1c^G
			// I know this is counterintuitive, but it's needed.
			time.Sleep(500 * time.Millisecond)

			projectDir, generationErr = project_generator.GenerateProject(projectName, selectedProject, selectedDatabase, allowTestCases)
		}).
		Run()

	if spinnerErr != nil {
		fmt.Printf("error create spinner. Error: %s\n", spinnerErr)
		return
	}

	if generationErr != nil {
		fmt.Println(cancelledText.Render("Error making project: " + generationErr.Error()))
		return
	}

	fmt.Println(successText.Render("\nProject scaffolded successfully!"))
	fmt.Println(infoText.Render("\nPlease cd into your project. Project location: " + projectDir))

}
