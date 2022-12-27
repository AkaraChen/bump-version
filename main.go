package main

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AkaraChen/bump-version/pkg/structs"
	"github.com/AkaraChen/bump-version/pkg/util"
	"github.com/Masterminds/semver/v3"

	"github.com/pterm/pterm"
	"github.com/tidwall/sjson"
)

func fileExist(pattern string) bool {
	file, err := filepath.Glob(pattern)
	if err != nil {
		return false
	}
	return len(file) != 0
}

var (
	execPath      = getExecAbsolutePath()
	__DEV__       = strings.Contains(execPath, "Temp") || strings.Contains(execPath, "tmp") || strings.Contains(execPath, "var/folders")
	pathSeparator = string(os.PathSeparator)
	cwd, _        = os.Getwd()
	currentPath   = strings.Join([]string{cwd, pathSeparator}, "")
	__MONOREPO__  = false
	mainPackage   string
	packages      []string
	config        structs.Config
)

func confirm(question string) bool {
	result, _ := pterm.
		DefaultInteractiveConfirm.
		Show(question)
	return result
}

func GetBumpedArray(version semver.Version) []string {
	major := version.IncMajor()
	minor := version.IncMinor()
	patch := version.IncPatch()
	return []string{"Major " + major.Original(), "Minor " + minor.Original(), "Patch " + patch.Original()}
}

func main() {
	pkg := structs.GetPackage(mainPackage)
	oldVersion, _ := semver.NewVersion(pkg.Version)
	versionArr := GetBumpedArray(*oldVersion)
	bumpType, _ := pterm.DefaultInteractiveSelect.
		WithOptions(versionArr).
		Show("Select release type:")
	newVersion := *oldVersion
	switch util.FindIndex(versionArr, bumpType) {
	case 0:
		newVersion = newVersion.IncMajor()
	case 1:
		newVersion = newVersion.IncMinor()
	case 2:
		newVersion = newVersion.IncPatch()
	}
	versionString := newVersion.Original()
	for _, file := range packages {
		bytes, _ := os.ReadFile(file)
		result, _ := sjson.SetBytes(bytes, "version", versionString)
		os.WriteFile(file, result, fs.ModeDevice)
	}
	if config.PrePublish != "" {
		run("pnpm", "run", config.PrePublish)
	}
	pterm.Info.Printfln("Generate changelog...")
	if config.Changelog {
		run("conventional-changelog", "-p", "angular", "-i", "CHANGELOG.md", "-s")
	}
	run("git", "add", ".")
	run("git", "commit", "-m", "release: "+versionString)
	if config.WithTag {
		run("git", "tag", versionString)
	}
	pterm.Info.Printfln("Push your change...")
	run("git", "push")
	if config.Publish && confirm("Would you like to publish to npm?") {
		cmd := exec.Command("npm", "publish")
		cmd.Stdin = os.Stdin
		cmd.Dir = currentPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}

}
