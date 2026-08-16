package project_generator

import "testing"

func TestValidateProjectNameAccepts(t *testing.T) {
	for _, name := range []string{
		"funApi",
		"my-cli",
		"my_cli",
		"api.v2",
		"a",
		"Project123",
	} {
		if err := ValidateProjectName(name); err != nil {
			t.Errorf("ValidateProjectName(%q) = %v, want nil", name, err)
		}
	}
}

func TestValidateProjectNameRejects(t *testing.T) {
	for _, name := range []string{
		"",           // nothing entered
		" ",          // whitespace only
		"my project", // breaks go mod init
		" api",       // leading space
		"api ",       // trailing space
		"../escape",  // would climb out of the output directory
		"nested/api", // would land a level down
		`nested\api`, // the same on windows
		".",          // the output directory itself
		"..",         // its parent
		".hidden",    // hidden directory
		"-flagish",   // reads as a flag on the command line
		"con",        // reserved on windows
		"NUL",        // reserved, any case
		"lpt1.txt",   // reserved, extension does not help
		"api:8080",   // rejected by go mod init
		"emoji🎉",     // outside the allowed set
	} {
		if err := ValidateProjectName(name); err == nil {
			t.Errorf("ValidateProjectName(%q) = nil, want an error", name)
		}
	}
}

func TestValidateProjectNameLength(t *testing.T) {
	long := make([]byte, maxProjectNameLen+1)
	for i := range long {
		long[i] = 'a'
	}
	if err := ValidateProjectName(string(long)); err == nil {
		t.Error("an over-long name was accepted, want an error")
	}
}
