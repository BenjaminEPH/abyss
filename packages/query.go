package packages

import (
	"log"
	"os/exec"
	"strings"
)

type Package struct {
	status      string
	name        string
	description string
}

func queryInstalledPackages() string {
	out, err := exec.Command("xbps-query", "-l").Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(out)
}

func splitInstalledPackages(packages string) []Package {
	var vec []Package
	for line := range strings.Lines(packages) {
		sep := strings.Fields(line)
		vec = append(vec, Package{
			status:      sep[0],
			name:        sep[1],
			description: strings.Join(sep[2:], " "),
		})

	}
	return vec
}
