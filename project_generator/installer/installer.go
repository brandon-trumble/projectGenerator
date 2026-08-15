// Package installer copies the running binary into a user-local bin directory
// and makes sure that directory is on the user's PATH, so the tool can be
// launched by name from any shell after the first run.
package installer

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"charm.land/huh/v2"
)

// CommandName is the name the binary takes once it lives on the user's PATH.
const CommandName = "project-gen"

// marker tags the lines we append to shell startup files so we only ever add
// them once, and so the user can see where they came from.
const marker = "# added by " + CommandName

// Result describes what the setup step actually did, so the caller can report it.
type Result struct {
	Dir        string   // the bin directory we install into
	BinaryPath string   // full path of the installed binary
	Copied     bool     // the binary was copied this run
	PathEdited bool     // PATH was changed (needs a new shell to take effect)
	ShellFiles []string // startup files we appended to
	AlreadySet bool     // nothing to do, install is already complete
	Declined   bool     // user said no, now or on a previous run
}

// MaybeSetup installs the binary and fixes up PATH, asking first. It is a no-op
// when the tool is already installed, when running through "go run", or when the
// user has declined before.
func MaybeSetup() (Result, error) { return setup(false) }

// Setup runs the same install, but ignores an earlier "not now" answer. This is
// what the --install flag calls.
func Setup() (Result, error) { return setup(true) }

func setup(force bool) (Result, error) {
	var res Result

	exe, err := currentBinary()
	if err != nil {
		return res, err
	}

	// "go run" builds into a temp dir, so there is nothing worth installing
	// unless the user explicitly asked for it
	if !force && isTempBuild(exe) {
		return res, nil
	}

	dir, err := binDir()
	if err != nil {
		return res, err
	}
	res.Dir = dir
	res.BinaryPath = filepath.Join(dir, binaryFileName())

	installed := sameFile(exe, res.BinaryPath)
	if installed && configured(dir) {
		res.AlreadySet = true
		return res, nil
	}

	if !force && declinedBefore() {
		res.Declined = true
		return res, nil
	}

	var confirmed bool
	err = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Install " + CommandName + " so you can run it from anywhere?").
				Description("Copies this binary to " + prettyPath(dir) + " and adds it to your PATH.").
				Affirmative("Yes").
				Negative("Not now").
				Value(&confirmed),
		),
	).Run()
	if err != nil {
		// a cancelled prompt just means "skip setup", the generator still works
		return res, nil
	}

	if !confirmed {
		rememberDecline()
		res.Declined = true
		return res, nil
	}

	if !installed {
		if err := copyBinary(exe, res.BinaryPath); err != nil {
			return res, err
		}
		res.Copied = true
	}

	edited, files, err := ensureOnPath(dir)
	if err != nil {
		return res, err
	}
	res.PathEdited = edited
	res.ShellFiles = files
	clearDecline()

	return res, nil
}

// currentBinary resolves the path of the running executable, following symlinks
// so a linked binary is not mistaken for a separate install.
func currentBinary() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not locate the running binary: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return filepath.Clean(exe), nil
}

func isTempBuild(exe string) bool {
	return strings.Contains(exe, "go-build") || strings.HasPrefix(exe, filepath.Clean(os.TempDir())+string(os.PathSeparator))
}

func binaryFileName() string {
	if runtime.GOOS == "windows" {
		return CommandName + ".exe"
	}
	return CommandName
}

// binDir picks the conventional per-user bin directory for the platform. Both
// are writable without admin rights, which keeps the install prompt-free.
func binDir() (string, error) {
	if runtime.GOOS == "windows" {
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			return filepath.Join(local, "Programs", CommandName), nil
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not find your home directory: %w", err)
	}

	if runtime.GOOS == "windows" {
		return filepath.Join(home, "AppData", "Local", "Programs", CommandName), nil
	}
	return filepath.Join(home, ".local", "bin"), nil
}

func sameFile(a, b string) bool {
	aInfo, err := os.Stat(a)
	if err != nil {
		return false
	}
	bInfo, err := os.Stat(b)
	if err != nil {
		return false
	}
	return os.SameFile(aInfo, bInfo)
}

// onPath reports whether dir is already one of the entries in $PATH.
func onPath(dir string) bool {
	dir = filepath.Clean(dir)
	for _, entry := range filepath.SplitList(os.Getenv("PATH")) {
		if entry == "" {
			continue
		}
		entry = filepath.Clean(os.ExpandEnv(entry))
		if entry == dir || (runtime.GOOS == "windows" && strings.EqualFold(entry, dir)) {
			return true
		}
	}
	return false
}

// copyBinary writes the running binary to dest. It stages the copy next to the
// destination and renames it, so replacing an older copy that is currently
// running does not fail with "text file busy".
func copyBinary(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("could not create %s: %w", filepath.Dir(dest), err)
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("could not read the running binary: %w", err)
	}
	defer in.Close()

	staged := dest + ".new"
	out, err := os.OpenFile(staged, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("could not write to %s: %w", filepath.Dir(dest), err)
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(staged)
		return fmt.Errorf("could not copy the binary: %w", err)
	}
	if err := out.Close(); err != nil {
		os.Remove(staged)
		return fmt.Errorf("could not copy the binary: %w", err)
	}

	if err := os.Rename(staged, dest); err != nil {
		os.Remove(staged)
		return fmt.Errorf("could not install to %s: %w", dest, err)
	}
	return nil
}

// configured reports whether dir will be on PATH in a shell started from now on.
// A PATH edit does not reach the running process, so right after an install this
// is true while onPath is still false.
func configured(dir string) bool {
	return onPath(dir) || persisted(dir)
}

// persisted reports whether the PATH edit is already written to disk.
func persisted(dir string) bool {
	if runtime.GOOS == "windows" {
		entries, err := userPathWindows()
		if err != nil {
			return false
		}
		for _, entry := range entries {
			if strings.EqualFold(filepath.Clean(entry), filepath.Clean(dir)) {
				return true
			}
		}
		return false
	}

	for _, rc := range shellFiles() {
		if body, err := os.ReadFile(rc); err == nil && strings.Contains(string(body), marker) {
			return true
		}
	}
	return false
}

// ensureOnPath makes dir permanently reachable from a fresh shell. It reports
// whether anything changed and which files were touched.
func ensureOnPath(dir string) (bool, []string, error) {
	if onPath(dir) {
		return false, nil, nil
	}
	if runtime.GOOS == "windows" {
		changed, err := ensurePathWindows(dir)
		return changed, nil, err
	}
	return ensurePathUnix(dir)
}

// userPathWindows reads the per-user Path environment variable.
func userPathWindows() ([]string, error) {
	ps, err := exec.LookPath("powershell")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(ps, "-NoProfile", "-NonInteractive", "-Command",
		"[Environment]::GetEnvironmentVariable('Path', 'User')")

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return filepath.SplitList(strings.TrimSpace(string(out))), nil
}

// ensurePathWindows appends dir to the per-user Path environment variable.
// User scope means no admin rights and no system-wide edit.
func ensurePathWindows(dir string) (bool, error) {
	ps, err := exec.LookPath("powershell")
	if err != nil {
		return false, fmt.Errorf("powershell not found, add %s to your PATH manually", dir)
	}

	// the directory travels in an env var so it never has to be quoted into the script
	const script = `
$dir = $env:PROJECT_GEN_BIN_DIR
$current = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($null -eq $current) { $current = '' }
$parts = @($current -split ';' | Where-Object { $_ -ne '' })
if ($parts -contains $dir) { Write-Output 'unchanged'; exit 0 }
$parts += $dir
[Environment]::SetEnvironmentVariable('Path', ($parts -join ';'), 'User')
Write-Output 'changed'
`

	cmd := exec.Command(ps, "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.Env = append(os.Environ(), "PROJECT_GEN_BIN_DIR="+dir)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("could not update your PATH: %s", strings.TrimSpace(string(out)))
	}
	return strings.Contains(string(out), "changed"), nil
}

// ensurePathUnix appends a PATH line to the startup files of the user's shell.
func ensurePathUnix(dir string) (bool, []string, error) {
	var touched []string

	for _, rc := range shellFiles() {
		added, err := appendOnce(rc, pathLine(rc, dir))
		if err != nil {
			return len(touched) > 0, touched, err
		}
		if added {
			touched = append(touched, rc)
		}
	}

	if len(touched) == 0 {
		// nothing to add is fine when a previous run already did it
		if persisted(dir) {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("could not find a shell startup file, add %s to your PATH manually", dir)
	}
	return true, touched, nil
}

// shellFiles lists the startup files worth editing for the current shell.
// ~/.profile is always included so login shells and desktop launchers see the
// change too.
func shellFiles() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	shell := filepath.Base(os.Getenv("SHELL"))
	var files []string

	switch shell {
	case "zsh":
		files = append(files, filepath.Join(home, ".zshrc"))
	case "fish":
		files = append(files, filepath.Join(home, ".config", "fish", "config.fish"))
	case "bash":
		files = append(files, filepath.Join(home, ".bashrc"))
	default:
		// unknown shell, fall back to the files bash and sh read
		files = append(files, filepath.Join(home, ".bashrc"))
	}

	if shell != "fish" {
		files = append(files, filepath.Join(home, ".profile"))
	}
	return files
}

func pathLine(rc, dir string) string {
	target := prettyPath(dir)
	if filepath.Ext(rc) == ".fish" {
		return marker + "\nfish_add_path " + target + "\n"
	}
	return marker + "\nexport PATH=\"" + target + ":$PATH\"\n"
}

// prettyPath rewrites a path inside the home directory to use $HOME, which
// keeps the shell config portable and readable.
func prettyPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || runtime.GOOS == "windows" {
		return path
	}
	if rel, err := filepath.Rel(home, path); err == nil && !strings.HasPrefix(rel, "..") {
		return "$HOME/" + filepath.ToSlash(rel)
	}
	return path
}

// appendOnce adds line to rc unless our marker is already in there. Missing
// files are created, missing parent directories are not.
func appendOnce(rc, line string) (bool, error) {
	existing, err := os.ReadFile(rc)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("could not read %s: %w", rc, err)
	}
	if strings.Contains(string(existing), marker) {
		return false, nil
	}
	if os.IsNotExist(err) {
		// only create a startup file in a directory that already exists,
		// creating ~/.config/fish for a shell the user does not have is noise
		if _, statErr := os.Stat(filepath.Dir(rc)); statErr != nil {
			return false, nil
		}
	}

	f, err := os.OpenFile(rc, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return false, fmt.Errorf("could not open %s: %w", rc, err)
	}
	defer f.Close()

	prefix := "\n"
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		prefix = "\n\n"
	}
	if _, err := f.WriteString(prefix + line); err != nil {
		return false, fmt.Errorf("could not write to %s: %w", rc, err)
	}
	return true, nil
}

// declineMarker is the file that remembers a "not now" answer, so the prompt
// only shows up once.
func declineMarker() (string, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfg, CommandName, "skip-install"), nil
}

func declinedBefore() bool {
	path, err := declineMarker()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

func rememberDecline() {
	path, err := declineMarker()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return
	}
	// best effort, a failure here only means the prompt shows again next run
	_ = os.WriteFile(path, []byte("run with --install to set this up later\n"), 0o600)
}

func clearDecline() {
	if path, err := declineMarker(); err == nil {
		_ = os.Remove(path)
	}
}
