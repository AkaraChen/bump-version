package main

import (
	"github.com/AkaraChen/bump-version/pkg/structs"
	"github.com/AkaraChen/bump-version/pkg/util"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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
		configByte, _ := ioutil.ReadFile(configDir)
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
		pterm.Printfln("You have uncommited change, would you like to continue without commit?")
		result, _ := pterm.DefaultInteractiveConfirm.Show()
		if !result {
			os.Exit(1)
		}
	}
}

func publish() {
	publish := exec.Command("npm", "publish")
	publish.Dir = currentPath
	publish.Stdout = os.Stdout
	publish.Stderr = os.Stderr
	publish.Stdin = os.Stdin
	publish.Run()
}

func init(){
	checkEnv()
	setup()
	checkGitStatus()
}

func main() {
	pkg := structs.GetPackage(mainPackage)
	oldVersion := structs.ParseVersion(pkg.Version)
	bumpType, _ := pterm.DefaultInteractiveSelect.
		WithOptions(oldVersion.GetBumpedArray()).
		Show("Select release type:")
	selectIndex := util.FindIndex(oldVersion.GetBumpedArray(), bumpType)
	newVersion := oldVersion.Bump(structs.VersionEnum(selectIndex))
	versionString := newVersion.ToString()
	for _, file := range packages {
		bytes, _ := ioutil.ReadFile(file)
		result, _ := sjson.SetBytes(bytes, "version", versionString)
		os.WriteFile(file, result, fs.ModeDevice)
	}
	pterm.Info.Printfln("Generate changelog...")
	run("conventional-changelog", "-p angular -i CHANGELOG.md -s")
	repo, err := git.PlainOpen(currentPath)
	if err != nil {
		pterm.Error.Printfln("Git repo not found in current dir.")
	}
	workTree, _ := repo.Worktree()
	workTree.Add(".")
	workTree.Commit("release: "+versionString, &git.CommitOptions{})
	repo.CreateTag(versionString, plumbing.NewHash(versionString), &git.CreateTagOptions{})
	pterm.Info.Printfln("Push your change...")
	repo.Push(&git.PushOptions{})
	confirmPublish, _ := pterm.
		DefaultInteractiveConfirm.
		Show("Would you like to publish to npm?")
	if confirmPublish {
		publish()
	}
}
