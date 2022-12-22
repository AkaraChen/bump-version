package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mattn/go-zglob"
	"github.com/pterm/pterm"
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
	__DEV__       = strings.Contains(execPath, "Temp") || strings.Contains(execPath, "tmp")
	pathSeparator = string(os.PathSeparator)
	cwd, _        = os.Getwd()
	currentPath   = strings.Join([]string{cwd, pathSeparator}, "")
)

func main() {
	if __DEV__ {
		pterm.Info.Println("Currently in development mode.")
		currentPath = strings.Join([]string{currentPath, "test"}, "")
		pterm.Info.Printfln("Current path: %s.", pterm.Blue(currentPath))
	}

	__MONOREPO__ := fileExist(strings.Join([]string{currentPath, pathSeparator, "pnpm-workspace.yaml"}, ""))

	mainPackage := strings.Join([]string{currentPath, pathSeparator, "package.json"}, "")
	if !fileExist(mainPackage) {
		pterm.Error.Printfln("Can't find package.json.")
		os.Exit(1)
	}

	packages := []string{mainPackage}
	if __MONOREPO__ {
		packagesPattern := strings.Join([]string{currentPath, pathSeparator, "**", pathSeparator, "package.json"}, "")
		result, err := zglob.Glob(packagesPattern)
		if err != nil || len(packages) == 0 {
			pterm.Error.Printfln("Can't find packages.")
		}
		packages = result
	}
	if __DEV__ && __MONOREPO__ {
		pterm.Info.Printfln("Monorepo founded.")
		pterm.Info.Printfln("Find %s packages in workspace.", pterm.Blue(len(packages)))
	}
}
