package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ProfileRule struct {
	MethodRegex *regexp.Regexp
	UrlRegex    *regexp.Regexp
}

func getProfile(profileName string) (rules []ProfileRule) {
	profilePath := filepath.Join("/profiles/", profileName)

	if _, err := os.Stat(profilePath); err != nil {
		log.Fatal("Could not read profile file: ", err)
	}

	profileHandle, err := os.Open(profilePath)
	if err != nil {
		log.Fatal("Error opening profile file: ", err)
	}
	defer profileHandle.Close()

	scanner := bufio.NewScanner(profileHandle)
	configFormatRegex := regexp.MustCompile(`^(\S+)\s+(/.+)$`)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "#") && line != "" {
			parsed := configFormatRegex.FindStringSubmatch(line)
			if parsed == nil {
				log.Fatal("Profile rule is invalid: ", line)
			}
			methodRegex, err := regexp.Compile("^" + strings.TrimSuffix(strings.TrimPrefix(parsed[1], "^"), "$") + "$")
			if err != nil {
				log.Fatal("Profile rule is invalid: ", err)
			}
			urlRegex, err := regexp.Compile("^" + strings.TrimSuffix(strings.TrimPrefix(parsed[2], "^"), "$") + "$")
			if err != nil {
				log.Fatal("Profile rule is invalid: ", err)
			}
			rules = append(rules, ProfileRule{MethodRegex: methodRegex, UrlRegex: urlRegex})
		}
	}
	if scanner.Err() != nil {
		log.Fatal("Error reading profile file: ", scanner.Err())
	}
	if len(rules) == 0 {
		log.Fatal("Profile file doesn't contain rules. Aborting.")
	}
	return
}
