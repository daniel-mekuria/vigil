//go:build !windows

package collector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func diskUsage(path string) (float64, float64, error) {
	if strings.TrimSpace(path) == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return 0, 0, err
		}
	}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(filepath.Clean(path), &stat); err != nil {
		return 0, 0, err
	}
	total := float64(stat.Blocks) * float64(stat.Bsize)
	available := float64(stat.Bavail) * float64(stat.Bsize)
	if total <= 0 || available < 0 || available > total {
		return 0, 0, fmt.Errorf("invalid filesystem capacity")
	}
	return total - available, total, nil
}
