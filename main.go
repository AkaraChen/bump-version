package main

import (
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AkaraChen/bump-version/pkg/structs"
	"github.com/AkaraChen/bump-version/pkg/util"
	"github.com/Masterminds/semver/v3"

	"github.com/mattn/go-zglob"
	"github.com/pterm/pterm"
	"github.com/tidwall/sjson"
	"gopkg.in/yaml.v2"
)

func getExecAbsolutePath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))
	return path[:index]
}

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
)

type Workspace struct {
	Packages []string `yaml:"packages"`
}

func checkPnpmWorkspace() {
	configDir := filepath.Join(currentPath, "pnpm-workspace.yaml")
	__MONOREPO__ = fileExist(configDir)
	if __MONOREPO__ {
		configByte, _ := os.ReadFile(configDir)
		var config Workspace
		err := yaml.Unmarshal(configByte, &config)
		if err != nil {
			pterm.Error.Printfln("Invalid workspace config.")
			os.Exit(1)
		}
		if len(config.Packages) > 0 {
			for _, value := range config.Packages {
				pattern := strings.Join([]string{currentPath, pathSeparator, value, pathSeparator, "**", pathSeparator, "package.json"}, "")
				result, _ := zglob.Glob(pattern)
				packages = append(packages, result...)
			}
		} else {
			pterm.Warning.Printfln("Workspace config founded, but have no packages field.")
		}
	}
}

func setup() {
	if __DEV__ {
		pterm.Info.Println("Currently in development mode.")
		currentPath = filepath.Join(cwd, "test")
	}

	pterm.Info.Printfln("Current path: %s.", pterm.Blue(currentPath))
	mainPackage = filepath.Join(currentPath, "package.json")
	if !fileExist(mainPackage) {
		pterm.Error.Printfln("Can't find package.json.")
		os.Exit(1)
	}

	packages = []string{mainPackage}
	checkPnpmWorkspace()

	if __DEV__ && __MONOREPO__ {
		pterm.Info.Printfln("Monorepo founded.")
		pterm.Info.Printfln("Find %s packages in workspace.", pterm.Blue(len(packages)))
	}
}

func run(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = currentPath
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	return string(output)
}

func checkProgram(name string) {
	_, err := exec.LookPath(name)
	if err != nil {
		pterm.Error.Printfln("%s not found in your environment.", name)
		os.Exit(1)
	}
}

func checkEnv() {
	checkProgram("node")
	checkProgram("npm")
	checkProgram("git")
	checkProgram("conventional-changelog")
}

func checkGitStatus() {
	if run("git", "status", "--porcelain") != "" && !__DEV__ {
		if !confirm("You have uncommited change, would you like to continue without commit?") {
			os.Exit(1)
		}
	}
}

func init() {
	checkEnv()
	setup()
	checkGitStatus()
}

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
	if !__DEV__ {
		pterm.Info.Printfln("Generate changelog...")
		run("conventional-changelog", "-p", "angular", "-i", "CHANGELOG.md", "-s")
		run("git", "add", ".")
		run("git", "commit", "-m", "release: "+versionString)
		run("git", "tag", versionString)
		pterm.Info.Printfln("Push your change...")
		run("git", "push")
		if confirm("Would you like to publish to npm?") {
			run("npm", "publish")
		}
	}

}
