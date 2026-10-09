package main

import (
	"fmt"
	"strings"
)

func greet(name string) string {
	if trimmedName := strings.TrimSpace(name); trimmedName != "" {

		return fmt.Sprintf("Hello, %s!", trimmedName)
	} else {
		return "Hello, World!"
	}
}
