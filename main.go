package main

import (
	"errors"
	"fmt"
	"os"
	"projectGenerator/project_generator"
	"projectGenerator/project_generator/forms"
	"projectGenerator/project_generator/installer"
	"runtime"
	"strings"
	"time"

	"charm.land/huh/v2"
	"charm.land/huh/v2/spinner"
	"charm.land/lipgloss/v2"
)

// stamped into release builds by goreleaser through -ldflags. A build straight
// from source leaves them at these defaults.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {

	// there are only two flags, so a comparison beats wiring up the flag package.
	// this accepts install, -install and --install alike.
	var flagArg string
	if len(os.Args) > 1 {
		flagArg = strings.TrimLeft(os.Args[1], "-")
	}

	if flagArg == "version" {
		fmt.Println(versionString())
		return
	}

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

	// on the first run, offer to drop the binary somewhere on the user's PATH so
	// it can be launched by name from any directory. --install redoes this on
	// demand, and skips straight to it.
	onlySetup := flagArg == "install"
	reportSetup(onlySetup, successText, infoText, cancelledText)
	if onlySetup {
		return
	}

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

			projectDir, generationErr = project_generator.GenerateProject(projectName, selectedProject, selectedDatabase, cliFramework, allowTestCases)
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

// versionString describes the running build, in the form goreleaser stamped it.
func versionString() string {
	if version == "dev" {
		return installer.CommandName + " dev (built from source)"
	}
	return fmt.Sprintf("%s %s (commit %s, built %s)", installer.CommandName, version, commit, date)
}

// reportSetup runs the PATH install and prints what it did. A failure here is
// never fatal, the generator still works when launched from its own directory.
func reportSetup(forced bool, successText, infoText, cancelledText lipgloss.Style) {
	var res installer.Result
	var err error

	if forced {
		res, err = installer.Setup()
	} else {
		res, err = installer.MaybeSetup()
	}

	if err != nil {
		fmt.Println(cancelledText.Render("Could not finish PATH setup: " + err.Error()))
		return
	}

	if res.AlreadySet {
		if forced {
			fmt.Println(infoText.Render("Already installed at " + res.BinaryPath))
		}
		return
	}

	// declined, skipped, or nothing to do
	if !res.Copied && !res.PathEdited {
		return
	}

	if res.Copied {
		fmt.Println(successText.Render("Installed to " + res.BinaryPath))
	}

	if res.PathEdited {
		for _, file := range res.ShellFiles {
			fmt.Println(infoText.Render("Added to your PATH in " + file))
		}
		if len(res.ShellFiles) == 0 {
			fmt.Println(infoText.Render("Added " + res.Dir + " to your PATH"))
		}
		fmt.Println(infoText.Render("Open a new terminal to pick it up, then run: " + installer.CommandName))
	} else {
		fmt.Println(infoText.Render("Run it from anywhere with: " + installer.CommandName))
	}

	fmt.Println()
}
