package project_generator

import (
	"fmt"
	"strings"
)

// maxProjectNameLen keeps the name short enough to stay a comfortable directory
// name on every platform.
const maxProjectNameLen = 64

// reservedNames are the device names Windows refuses to use as a directory,
// whatever the extension. Creating one fails with an error that says nothing
// about the name, so they are worth catching up front.
var reservedNames = map[string]bool{
	"con": true, "prn": true, "aux": true, "nul": true,
	"com1": true, "com2": true, "com3": true, "com4": true,
	"com5": true, "com6": true, "com7": true, "com8": true, "com9": true,
	"lpt1": true, "lpt2": true, "lpt3": true, "lpt4": true,
	"lpt5": true, "lpt6": true, "lpt7": true, "lpt8": true, "lpt9": true,
}

// ValidateProjectName reports whether name works as both a directory name and a
// module path for "go mod init". The name form calls it as the user types, and
// GenerateProject calls it again so a name that would write outside the output
// directory cannot reach the filesystem.
func ValidateProjectName(name string) error {
	switch {
	case name == "":
		return fmt.Errorf("please enter a project name")
	case strings.TrimSpace(name) != name:
		return fmt.Errorf("project name cannot start or end with a space")
	case len(name) > maxProjectNameLen:
		return fmt.Errorf("project name is longer than %d characters", maxProjectNameLen)
	case name == "." || name == "..":
		return fmt.Errorf("%q is not a project name", name)
	case strings.HasPrefix(name, "-"), strings.HasPrefix(name, "."):
		return fmt.Errorf("project name cannot start with %q", name[:1])
	case reservedNames[strings.ToLower(strings.SplitN(name, ".", 2)[0])]:
		return fmt.Errorf("%q is a reserved device name on Windows, please pick another", name)
	}

	// "go mod init" rejects most punctuation, and a separator would put the
	// project somewhere other than the directory we are about to report
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '-', r == '_', r == '.':
		case r == ' ':
			return fmt.Errorf("project name cannot contain spaces, try %q", strings.ReplaceAll(name, " ", "-"))
		default:
			return fmt.Errorf("project name cannot contain %q, use letters, digits, '-', '_' or '.'", string(r))
		}
	}

	return nil
}
