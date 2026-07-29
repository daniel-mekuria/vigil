//go:build windows

package collector

import "fmt"

func diskUsage(string) (float64, float64, error) {
	return 0, 0, fmt.Errorf("unsupported platform windows")
}
