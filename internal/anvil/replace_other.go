//go:build !windows

package anvil

import "os"

func replaceRegionFile(source, target string) error {
	return os.Rename(source, target)
}
