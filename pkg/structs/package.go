package structs

import (
	"encoding/json"
	"io/ioutil"
)

type Package struct {
	Version string `json:"version"`
	Name    string `json:"name"`
}

func GetPackage(file string) Package {
	byte, _ := ioutil.ReadFile(file)
	var pkg Package
	json.Unmarshal(byte, &pkg)
	return pkg
}
