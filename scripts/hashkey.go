//go:build ignore

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter API key: ")
	key, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	key = strings.TrimSpace(key)
	if key == "" {
		fmt.Fprintln(os.Stderr, "API key cannot be empty")
		os.Exit(1)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating hash: %v\n", err)
		os.Exit(1)
	}

	// For Makefile use, $ must be escaped as $$
	escaped := strings.ReplaceAll(string(hash), "$", "$$")
	fmt.Printf("\nAPI Key Hash (add to local.mk):\n")
	fmt.Printf("export API_KEY_HASH=%s\n", escaped)
}
