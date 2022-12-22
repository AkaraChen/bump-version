package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mattn/go-zglob"
	"github.com/pterm/pterm"
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

type Package struct {
	Version int `json:"version"`
}

func checkPnpmWorkspace() {
	workspaceConfigDir := filepath.Join(currentPath, "pnpm-workspace.yaml")
	__MONOREPO__ = fileExist(workspaceConfigDir)
	workspaceConfigByte, _ := ioutil.ReadFile(workspaceConfigDir)
	var workspaceConfig Workspace
	err := yaml.Unmarshal(workspaceConfigByte, &workspaceConfig)
	if err != nil {
		pterm.Error.Printfln("Invalid workspace config.")
		os.Exit(1)
	}
	if len(workspaceConfig.Packages) > 0 {
		for _, value := range workspaceConfig.Packages {
			packagesPattern := strings.Join([]string{currentPath, pathSeparator, value, pathSeparator, "**", pathSeparator, "package.json"}, "")
			result, _ := zglob.Glob(packagesPattern)
			packages = append(packages, result...)
		}
	} else {
		pterm.Warning.Printfln("Workspace config founded, but have no packages field.")
	}
}

func setup() {
	if __DEV__ {
		pterm.Info.Println("Currently in development mode.")
		currentPath = filepath.Join(cwd, "test")
		pterm.Info.Printfln("Current path: %s.", pterm.Blue(currentPath))
	}

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

func main() {
	setup()
}
