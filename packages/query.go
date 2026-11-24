package packages

import (
	"os/exec"
	"strings"

	"github.com/BenjaminEPH/abyss/types"
)

// QueryInstalled returns the packages reported by `xbps-query -l`.
func QueryInstalled() ([]types.Package, error) {
	out, err := exec.Command("xbps-query", "-l").Output()
	if err != nil {
		return nil, err
	}
	return splitInstalledPackages(string(out)), nil
}

// splitInstalledPackages parses `xbps-query -l` output, where each line is:
//
//	<status> <name>-<version>_<revision> <description...>
func splitInstalledPackages(output string) []types.Package {
	var pkgs []types.Package
	for line := range strings.Lines(output) {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		pkgs = append(pkgs, types.Package{
			Status:      fields[0],
			Name:        pkgName(fields[1]),
			Description: strings.Join(fields[2:], " "),
		})
	}
	return pkgs
}

// pkgName strips the "-version_revision" suffix: "bash-5.2.37_1" -> "bash".
func pkgName(pkgver string) string {
	if i := strings.LastIndex(pkgver, "-"); i >= 0 {
		return pkgver[:i]
	}
	return pkgver
}
