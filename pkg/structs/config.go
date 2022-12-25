package structs

import (
	"encoding/json"

	"github.com/pterm/pterm"
)

type Config struct {
	PrePublish string `json:"prePublish"`
	Changelog  bool   `json:"changelog"`
	Publish    bool   `json:"publish"`
	WithTag    bool   `json:"withTag"`
}

func getConfig(bytes []byte) Config {
	var config Config
	err := json.Unmarshal(bytes, &config)
	if err != nil {
		pterm.Error.Printfln("Invalid Config.")
	}
	return config
}
