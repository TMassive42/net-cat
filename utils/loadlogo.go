package utils

import (
	"fmt"
	"os"
)
func LoadLogo() (string, error) {
	content, err := os.ReadFile("logo.txt")
	if err != nil {
		return "", fmt.Errorf("error reading logo file: %v", err)
	}
	return string(content), nil
}
