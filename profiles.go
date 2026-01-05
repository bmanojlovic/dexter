package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Profile struct {
	Name      string
	Path      string
	IsDir     bool
	IsPrivate bool
}

func profileExists(profilesPath, name string) bool {
	// Check if it's a numeric index
	if num, err := strconv.Atoi(name); err == nil {
		profiles := listProfiles(profilesPath, false)
		return num > 0 && num <= len(profiles)
	}

	// Check for exact name match
	profiles := listProfiles(profilesPath, false)
	for _, p := range profiles {
		if p.Name == name {
			return true
		}
	}

	// Check if it's a valid file path
	profilePath := filepath.Join(profilesPath, name)
	if _, err := os.Stat(profilePath); err == nil {
		return true
	}

	return false
}

func listProfiles(profilesPath string, private bool) []Profile {
	var profiles []Profile
	var dirs []Profile

	entries, err := os.ReadDir(profilesPath)
	if err != nil {
		return profiles
	}

	for _, entry := range entries {
		fullPath := filepath.Join(profilesPath, entry.Name())

		if entry.IsDir() {
			dirIsPrivate := hasPrivateMarker(fullPath)
			if (private && dirIsPrivate) || (!private && !dirIsPrivate) {
				dirs = append(dirs, Profile{
					Name:      entry.Name() + "/",
					Path:      fullPath,
					IsDir:     true,
					IsPrivate: dirIsPrivate,
				})
			}
		} else if !strings.HasSuffix(entry.Name(), "~") && hasAWSProfile(fullPath) {
			parentIsPrivate := hasPrivateMarker(profilesPath)
			if (private && parentIsPrivate) || (!private && !parentIsPrivate) {
				profiles = append(profiles, Profile{
					Name:      entry.Name(),
					Path:      fullPath,
					IsDir:     false,
					IsPrivate: parentIsPrivate,
				})
			}
		}
	}

	// Append directories at the end
	profiles = append(profiles, dirs...)
	return profiles
}

func hasPrivateMarker(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".private"))
	return err == nil
}

func hasAWSProfile(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "export ") {
			return true
		}
	}
	return false
}

func loadProfile(root, profile string, private bool) {
	if num, err := strconv.Atoi(profile); err == nil {
		profiles := listProfiles(root, private)
		if num > 0 && num <= len(profiles) {
			selected := profiles[num-1]
			if selected.IsDir {
				chooseProfile(selected.Path, private)
			} else {
				loadProfileFile(selected.Path)
			}
		}
		return
	}

	profilePath := filepath.Join(root, profile)
	if info, err := os.Stat(profilePath); err == nil {
		if info.IsDir() {
			chooseProfile(profilePath, private)
		} else {
			loadProfileFile(profilePath)
		}
	}
}

func loadProfileFile(path string) {
	// Read file and extract only export lines for parsing
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		return
	}

	// Create a temporary string with only export lines for godotenv
	var exportLines []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "export ") {
			// Remove "export " prefix for godotenv
			exportLines = append(exportLines, strings.TrimPrefix(trimmed, "export "))
		}
	}

	// Parse env vars
	envMap, err := godotenv.Unmarshal(strings.Join(exportLines, "\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing env vars: %v\n", err)
		return
	}

	var warnings []string

	// Check for file path variables
	if awsConfigFile, ok := envMap["AWS_CONFIG_FILE"]; ok {
		filePath := strings.ReplaceAll(awsConfigFile, "~", os.Getenv("HOME"))
		filePath = os.ExpandEnv(filePath)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			warnings = append(warnings, fmt.Sprintf("AWS_CONFIG_FILE not found: %s", filePath))
		}
	}

	if kubeconfig, ok := envMap["KUBECONFIG"]; ok {
		filePath := strings.ReplaceAll(kubeconfig, "~", os.Getenv("HOME"))
		filePath = os.ExpandEnv(filePath)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			warnings = append(warnings, fmt.Sprintf("KUBECONFIG not found: %s", filePath))
		}
	}

	// Display warnings to stderr
	if len(warnings) > 0 {
		fmt.Fprintln(os.Stderr, "\n⚠️  Warnings:")
		for _, warning := range warnings {
			fmt.Fprintf(os.Stderr, "  - %s\n", warning)
		}
		fmt.Fprintln(os.Stderr)
	}

	// Output shell commands to stdout
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			fmt.Println(trimmed)
		}
	}
}
