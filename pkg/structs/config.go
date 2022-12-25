package structs

type Config struct {
	PrePublish string `json:"prePublish"`
	Changelog  bool   `json:"changelog"`
	Publish    bool   `json:"publish"`
	WithTag    bool   `json:"withTag"`
}
