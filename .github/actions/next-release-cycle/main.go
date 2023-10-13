package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: next-release-cycle <branch-name>")
	}

	branchName := strings.TrimPrefix(os.Args[1], "release-")
	log.Println("Determing release cycle after", branchName)

	currentVersion, err := semver.NewVersion(branchName)
	if err != nil {
		return err
	}

	nextVersion := semver.New(
		currentVersion.Major(),
		currentVersion.Minor()+1,
		0,
		"",
		"",
	)
	nextVersionText := "v" + nextVersion.String()

	// Set the next cycle as a variable in the GitHub Actions environment.
	if githubOutputFile := os.Getenv("GITHUB_OUTPUT"); githubOutputFile != "" {
		setVariable := fmt.Sprintf("next-cycle=%s\n", nextVersionText)
		if err := appendFile(githubOutputFile, []byte(setVariable)); err != nil {
			return err
		}
	}

	fmt.Fprintln(os.Stdout, nextVersionText)
	return nil
}

// appendFile appends data to the file named name. If the file doesn't exist,
// an error is returned.
func appendFile(name string, data []byte) error {
	f, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
}
