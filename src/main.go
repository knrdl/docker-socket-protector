package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
)

func IsSocket(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.Mode().Type() == fs.ModeSocket
}

func main() {
	logRequests, err := strconv.ParseBool(os.Getenv("LOG_REQUESTS"))
	if err != nil {
		log.Fatal("Environment variable LOG_REQUESTS must be 'true' or 'false'.")
	}
	profileName := os.Getenv("PROFILE")

	if profileName == "" {
		log.Fatal("No profile given. Use 'unprotected' to allow all requests.")
	}
	if profileName == "unprotected" && !logRequests {
		log.Fatal("Aborting. Allowing all requests (as profile is 'unprotected') and not logging requests is not recommended.")
	}

	if !IsSocket("/var/run/docker.sock") {
		log.Fatal("No docker socket provided at '/var/run/docker.sock'.")
	}

	profileRules := getProfile(profileName)

	handler := &FilterProxy{
		Rules:       &profileRules,
		LogRequests: logRequests,
		Forwarder:   NewForwarder(),
	}

	if err := http.ListenAndServe(":2375", handler); err != nil {
		log.Fatal("error listing on port: ", err)
	}
}
