package structs

import (
	"encoding/json"
	"os"
)

type Package struct {
	Version string `json:"version"`
	Name    string `json:"name"`
}

func GetPackage(file string) Package {
	byte, _ := os.ReadFile(file)
	var pkg Package
	json.Unmarshal(byte, &pkg)
	return pkg
}
