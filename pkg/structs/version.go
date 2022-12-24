package structs

import (
	"os"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
)

type Version struct {
	major int
	minor int
	patch int
}

func ParseVersion(version string) Version {
	stringArray := strings.Split(version, ".")
	array := []int{}
	for _, value := range stringArray {
		value, err := strconv.Atoi(value)
		if err != nil {
			pterm.Error.Printfln("Pre release it not supported at this time.")
			os.Exit(1)
		}
		array = append(array, value)
	}
	return Version{major: array[0], minor: array[1], patch: array[2]}
}

type VersionEnum int

const (
	Major VersionEnum = 0
	Minor VersionEnum = 1
	Patch VersionEnum = 2
)

var VersionEnumArray = []string{"Major", "Minor", "Patch"}

func (version Version) Bump(bumpType VersionEnum) Version {
	switch bumpType {
	case Major:
		{
			version.major = version.major + 1
			version.minor = 0
			version.patch = 0
		}
	case Minor:
		{
			version.minor = version.minor + 1
			version.patch = 0
		}
	case Patch:
		{
			version.patch = version.patch + 1
		}
	}
	return version
}

func (version Version) GetBumped() []Version {
	return []Version{
		version.Bump(Major),
		version.Bump(Minor),
		version.Bump(Patch),
	}
}

func (version Version) ToString() string {
	return strings.Join([]string{
		strconv.Itoa(version.major),
		strconv.Itoa(version.minor),
		strconv.Itoa(version.patch)}, ".")
}

func (version Version) GetBumpedArray() []string {
	var arr = []string{}
	for index, value := range version.GetBumped() {
		text := strings.Join([]string{VersionEnumArray[index], value.ToString()}, " ")
		arr = append(arr, text)
	}
	return arr
}
