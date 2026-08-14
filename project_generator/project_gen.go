package project_generator

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

//go:embed templates
var templates embed.FS

// templateFile maps a template's path inside the embedded FS to where it
// should land in the generated project, relative to parentDir.
type templateFile struct {
	src  string
	dest string
}

type TemplateData struct {
	ProjectName string
	GoVersion   string

	// the readme templates branch on these to describe the project that was
	// actually generated
	Database     string
	CliFramework string
	WithTests    bool
}

// GenerateProject takes in fields from form, and inserts directories and files based of these.
func GenerateProject(projectName, projectType, selectedDatabase, cliFramework string, allowTestCases bool) (string, error) {
	userHomeDir, _ := os.UserHomeDir()
	parentDir := filepath.Join(userHomeDir, "generated_go_projects", projectName)

	// check if they have go installed first
	_, err := exec.LookPath("go")
	if err != nil {
		return "", fmt.Errorf("golang not installed. Please install golang first")
	}

	// get their go version and strip the word "go" from the front
	userGoVer := strings.ReplaceAll(runtime.Version(), "go", "")

	if err := os.MkdirAll(parentDir, 0750); err != nil {
		return "", fmt.Errorf("error making directory. error: %s", err)
	}

	cmd := exec.Command("go", "mod", "init", projectName)
	cmd.Dir = parentDir

	// bubble error up
	// for now I do not think I need the output of this....
	_, err = cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error during go mod init, please check if project exists already. error: %s", err)
	}

	dirPaths := []string{
		filepath.Join(parentDir, ".github", "workflows"),
		filepath.Join(parentDir, "cmd"),
	}

	// files copied byte for byte instead of being run through text/template.
	// the html templates need this: their "{{ }}" belongs to the generated
	// project's own renderer, not to us.
	var rawFiles []templateFile

	// initial files,
	// I might change this to include more base ones
	templateFiles := []templateFile{
		{
			"templates/template.gitignore.txt",
			".gitignore",
		},
		{
			"templates/template.github.workflow.ci.txt",
			filepath.Join(".github", "workflows", "ci.yml"),
		},
	}

	httpBackStdLibTemplates := "templates/http_backend_templates/standard_library/"
	cliTemplates := "templates/cli_templates/"

	// set when the chosen templates import packages outside the standard library
	needsTidy := false

	switch projectType {
	case "http_backend":
		dirPaths = append(dirPaths,
			filepath.Join(parentDir, "cmd", "api"),
			filepath.Join(parentDir, "internal"),
			filepath.Join(parentDir, "internal", "domain"),
			filepath.Join(parentDir, "internal", "service"),
			filepath.Join(parentDir, "internal", "handler"),
		)

		// the repository sits behind an interface, so everything above it is the
		// same whichever database was picked
		shared := httpBackStdLibTemplates + "shared/"

		templateFiles = append(templateFiles,
			templateFile{
				"templates/http_backend_templates/template.http_backend.makefile.txt",
				"makefile",
			},

			templateFile{
				"templates/http_backend_templates/template.http_backend.readme.md.txt",
				"README.md",
			},

			templateFile{
				shared + "template.http_backend.main.go.txt",
				filepath.Join("cmd", "api", "main.go"),
			},

			templateFile{
				shared + "template.http_backend.setup_routes.go.txt",
				filepath.Join("cmd", "api", "setupRoutes.go"),
			},

			templateFile{
				shared + "template.http_backend.handler.user.go.txt",
				filepath.Join("internal", "handler", "user.go"),
			},

			templateFile{
				shared + "template.http_backend.service.user.go.txt",
				filepath.Join("internal", "service", "user.go"),
			},

			templateFile{
				shared + "template.http_backend.domain.user.go.txt",
				filepath.Join("internal", "domain", "user.go"),
			},
		)

		// the tests stub the repository out, so they run without a database
		if allowTestCases {
			templateFiles = append(templateFiles,
				templateFile{
					shared + "template.http_backend.handler.user_test.go.txt",
					filepath.Join("internal", "handler", "user_test.go"),
				},

				templateFile{
					shared + "template.http_backend.service.user_test.go.txt",
					filepath.Join("internal", "service", "user_test.go"),
				},
			)
		}

		switch selectedDatabase {

		case "postgres":

			dirPaths = append(dirPaths,
				filepath.Join(parentDir, "internal", "repository", "postgres"),
				filepath.Join(parentDir, "internal", "database"),
			)

			postGres := httpBackStdLibTemplates + "with_database/postgres/"

			// the pgx driver is not in the standard library, so the generated
			// module needs its dependencies resolved before it will build
			needsTidy = true

			templateFiles = append(templateFiles,
				templateFile{
					postGres + "template.http_backend.env.txt",
					".env",
				},

				templateFile{
					postGres + "template.http_backend.run.go.txt",
					filepath.Join("cmd", "api", "run.go"),
				},

				templateFile{
					postGres + "template.http_backend.repo.postgres.user.go.txt",
					filepath.Join("internal", "repository", "postgres", "user.go"),
				},

				templateFile{
					postGres + "template.http_backend.database.postgres.go.txt",
					filepath.Join("internal", "database", "postgres.go"),
				},

				// nothing creates the users table the repository reads from,
				// so ship the schema next to the connection code
				templateFile{
					postGres + "template.http_backend.schema.postgres.sql.txt",
					filepath.Join("internal", "database", "schema.sql"),
				},
			)

		case "mysql":

			dirPaths = append(dirPaths,
				filepath.Join(parentDir, "internal", "repository", "mysql"),
				filepath.Join(parentDir, "internal", "database"),
			)

			mySql := httpBackStdLibTemplates + "with_database/mysql/"

			// database/sql is standard library, but its mysql driver is not
			needsTidy = true

			templateFiles = append(templateFiles,
				templateFile{
					mySql + "template.http_backend.env.txt",
					".env",
				},

				templateFile{
					mySql + "template.http_backend.run.go.txt",
					filepath.Join("cmd", "api", "run.go"),
				},

				templateFile{
					mySql + "template.http_backend.repo.mysql.user.go.txt",
					filepath.Join("internal", "repository", "mysql", "user.go"),
				},

				templateFile{
					mySql + "template.http_backend.database.mysql.go.txt",
					filepath.Join("internal", "database", "mysql.go"),
				},

				// nothing creates the users table the repository reads from,
				// so ship the schema next to the connection code
				templateFile{
					mySql + "template.http_backend.schema.mysql.sql.txt",
					filepath.Join("internal", "database", "schema.sql"),
				},
			)

		// either none or DB that does not exist form template.
		// these get an in memory repository so the project still runs.
		default:

			dirPaths = append(dirPaths, filepath.Join(parentDir, "internal", "repository", "memory"))

			noDatabase := httpBackStdLibTemplates + "without_database/"

			templateFiles = append(templateFiles,
				templateFile{
					"templates/http_backend_templates/template.http_backend.env.txt",
					".env",
				},

				templateFile{
					noDatabase + "template.http_backend.run.go.txt",
					filepath.Join("cmd", "api", "run.go"),
				},

				templateFile{
					noDatabase + "template.http_backend.repo.memory.user.go.txt",
					filepath.Join("internal", "repository", "memory", "user.go"),
				},
			)
		}

	case "cli_project":
		// the entry point lives in cmd/<project name> so the built binary and
		// the package directory share a name
		cmdDir := filepath.Join("cmd", projectName)

		dirPaths = append(dirPaths,
			filepath.Join(parentDir, cmdDir),
			filepath.Join(parentDir, "internal"),
		)

		templateFiles = append(templateFiles,
			templateFile{
				"templates/cli_templates/template.cli.makefile.txt",
				"makefile",
			},

			templateFile{
				"templates/cli_templates/template.cli.readme.md.txt",
				"README.md",
			},
		)

		// every CLI framework here is a third party dependency
		needsTidy = true

		switch cliFramework {

		case "cli-cobra":
			cobra := cliTemplates + "cobra_framework/"

			dirPaths = append(dirPaths, filepath.Join(parentDir, "internal", "commands"))

			templateFiles = append(templateFiles,
				templateFile{
					cobra + "template.cli.main.go.txt",
					filepath.Join(cmdDir, "main.go"),
				},

				templateFile{
					cobra + "template.cli.commands.root.go.txt",
					filepath.Join("internal", "commands", "root.go"),
				},

				templateFile{
					cobra + "template.cli.commands.greet.go.txt",
					filepath.Join("internal", "commands", "greet.go"),
				},
			)

		case "cli-bubbletea":
			bubbleTea := cliTemplates + "bubbletea_framework/"

			dirPaths = append(dirPaths, filepath.Join(parentDir, "internal", "tui"))

			templateFiles = append(templateFiles,
				templateFile{
					bubbleTea + "template.cli.main.go.txt",
					filepath.Join(cmdDir, "main.go"),
				},

				templateFile{
					bubbleTea + "template.cli.model.go.txt",
					filepath.Join("internal", "tui", "model.go"),
				},
			)

		// huh is the default, so an unrecognized framework still scaffolds
		default:
			huh := cliTemplates + "huh_framework/"

			dirPaths = append(dirPaths, filepath.Join(parentDir, "internal", "prompt"))

			templateFiles = append(templateFiles,
				templateFile{
					huh + "template.cli.main.go.txt",
					filepath.Join(cmdDir, "main.go"),
				},

				templateFile{
					huh + "template.cli.prompt.go.txt",
					filepath.Join("internal", "prompt", "prompt.go"),
				},
			)
		}

	case "htmx_project":
		htmx := "templates/htmx_templates/"

		dirPaths = append(dirPaths,
			filepath.Join(parentDir, "cmd", "web"),
			filepath.Join(parentDir, "internal"),
			filepath.Join(parentDir, "internal", "domain"),
			filepath.Join(parentDir, "internal", "service"),
			filepath.Join(parentDir, "internal", "handler"),
			filepath.Join(parentDir, "internal", "repository", "memory"),
			filepath.Join(parentDir, "web", "templates"),
			filepath.Join(parentDir, "web", "static"),
		)

		templateFiles = append(templateFiles,
			templateFile{
				htmx + "template.htmx.makefile.txt",
				"makefile",
			},

			templateFile{
				htmx + "template.htmx.readme.md.txt",
				"README.md",
			},

			templateFile{
				"templates/http_backend_templates/template.http_backend.env.txt",
				".env",
			},

			// the startup, model, and in-memory store are the same ones the
			// http backend uses
			templateFile{
				httpBackStdLibTemplates + "shared/template.http_backend.main.go.txt",
				filepath.Join("cmd", "web", "main.go"),
			},

			templateFile{
				httpBackStdLibTemplates + "shared/template.http_backend.domain.user.go.txt",
				filepath.Join("internal", "domain", "user.go"),
			},

			templateFile{
				httpBackStdLibTemplates + "without_database/template.http_backend.repo.memory.user.go.txt",
				filepath.Join("internal", "repository", "memory", "user.go"),
			},

			templateFile{
				htmx + "template.htmx.run.go.txt",
				filepath.Join("cmd", "web", "run.go"),
			},

			templateFile{
				htmx + "template.htmx.setup_routes.go.txt",
				filepath.Join("cmd", "web", "setupRoutes.go"),
			},

			templateFile{
				htmx + "template.htmx.handler.page.go.txt",
				filepath.Join("internal", "handler", "page.go"),
			},

			templateFile{
				htmx + "template.htmx.service.user.go.txt",
				filepath.Join("internal", "service", "user.go"),
			},

			templateFile{
				htmx + "template.htmx.web.go.txt",
				filepath.Join("web", "web.go"),
			},
		)

		// the html keeps its own template syntax, so it is copied as is
		rawFiles = append(rawFiles,
			templateFile{
				htmx + "web/index.html",
				filepath.Join("web", "templates", "index.html"),
			},

			templateFile{
				htmx + "web/users.html",
				filepath.Join("web", "templates", "users.html"),
			},

			templateFile{
				htmx + "web/style.css",
				filepath.Join("web", "static", "style.css"),
			},
		)

	// an empty project still gets an entry point, otherwise there is nothing
	// for the generated CI workflow to build
	default:
		empty := "templates/empty_templates/"

		dirPaths = append(dirPaths, filepath.Join(parentDir, "cmd", projectName))

		templateFiles = append(templateFiles,
			templateFile{
				"templates/cli_templates/template.cli.makefile.txt",
				"makefile",
			},

			templateFile{
				empty + "template.empty.readme.md.txt",
				"README.md",
			},

			templateFile{
				empty + "template.empty.main.go.txt",
				filepath.Join("cmd", projectName, "main.go"),
			},
		)
	}

	for i := range dirPaths {
		// bugfix: for making .GitHub files,
		// I had a bug where they could not be made, so changing os.Mkdir to: os.MkdirAll fixed it.
		err := os.MkdirAll(dirPaths[i], 0750)
		if err != nil {
			return "", fmt.Errorf("error during internal file path creation. error: %s", err)
		}
	}

	// this will substitute the "{{ .ProjectName }}" in the text files.
	// it solves the internal scaffolding import problem I faced.
	templateData := TemplateData{
		ProjectName:  projectName,
		GoVersion:    userGoVer,
		Database:     selectedDatabase,
		CliFramework: cliFramework,
		WithTests:    allowTestCases,
	}

	for _, t := range templateFiles {
		rawData, err := templates.ReadFile(t.src)
		if err != nil {
			return "", fmt.Errorf("error during file template reading. error: %s", err)
		}

		tmpl, err := template.New(t.src).Parse(string(rawData))
		if err != nil {
			return "", fmt.Errorf("error parsing template %s. error: %s", t.src, err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, templateData); err != nil {
			return "", fmt.Errorf("error executing template %s. error: %s", t.src, err)
		}

		if err := os.WriteFile(filepath.Join(parentDir, t.dest), buf.Bytes(), 0660); err != nil {
			return "", fmt.Errorf("error during template file writing. error: %s", err)
		}
	}

	for _, r := range rawFiles {
		rawData, err := templates.ReadFile(r.src)
		if err != nil {
			return "", fmt.Errorf("error during raw file reading. error: %s", err)
		}

		if err := os.WriteFile(filepath.Join(parentDir, r.dest), rawData, 0660); err != nil {
			return "", fmt.Errorf("error during raw file writing. error: %s", err)
		}
	}

	if needsTidy {
		tidy := exec.Command("go", "mod", "tidy")
		tidy.Dir = parentDir

		if out, err := tidy.CombinedOutput(); err != nil {
			return "", fmt.Errorf("project files were written to %s, but downloading dependencies failed. please run go mod tidy there. error: %s: %s", parentDir, err, out)
		}
	}

	return parentDir, nil
}
