package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AkaraChen/bump-version/pkg/structs"
	"github.com/mattn/go-zglob"
	"github.com/pterm/pterm"
	"gopkg.in/yaml.v2"
)

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

func setUpConfig() {
	configDir := currentPath + pathSeparator + ".bumprc.json"
	if !fileExist(configDir) {
		if confirm("Bump config not found, would you like to create it ?") {
			_, err := os.Create(configDir)
			if err != nil {
				pterm.Error.Printfln("Cannot create config, pls check your permisson.")
			}
			byte, _ := json.Marshal(structs.Config{
				PrePublish: "build",
				Changelog:  true,
				Publish:    true,
				WithTag:    true,
			})
			var bytesBuffer bytes.Buffer
			json.Indent(&bytesBuffer, byte, "", "\t")
			os.WriteFile(configDir, bytesBuffer.Bytes(), os.ModeDevice)
			pterm.Printfln("Please checkout .bumprc.json, and run again!")
			os.Exit(0)
		}
	} else {
		var userConfig structs.Config
		bytes, _ := os.ReadFile(configDir)
		err := json.Unmarshal(bytes, &userConfig)
		if err != nil {
			pterm.Error.Printfln("Can't parse config.")
		}
		config = userConfig
	}
}

func getExecAbsolutePath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))
	return path[:index]
}

func init() {
	checkEnv()
	setup()
	checkGitStatus()
	setUpConfig()
}
