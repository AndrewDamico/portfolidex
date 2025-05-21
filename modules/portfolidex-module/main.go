package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type PortfolioEntry struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Owner       string   `json:"owner"`
	Description string   `json:"description"`
	DexModules  []string `json:"dex_modules"`
	Repo        string   `json:"repo"`
}

// Add this struct to match your .programidex/init_config.json structure
type ProgramidexConfig struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	GitHubRepo  string   `json:"github_repo"`
	GoModule    string   `json:"go_module"`
	Directories []string `json:"directories"`
	// Add other fields as needed
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	entry := PortfolioEntry{}

	// Try to read .programidex/init_config.json for defaults
	var defaultName, defaultType, defaultRepo string
	var defaultModules []string
	if cfg := readProgramidexConfig(); cfg != nil {
		defaultName = cfg.Name
		defaultType = cfg.Type
		defaultRepo = cfg.GitHubRepo
		defaultModules = getDexModulesFromConfig(cfg)
	}

	// Fallback to auto-detect project name if not in config
	if defaultName == "" {
		cwd, _ := os.Getwd()
		defaultName = filepath.Base(cwd)
	}

	fmt.Printf("Project/App Name [%s]: ", defaultName)
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		name = defaultName
	}
	entry.Name = name

	if defaultType == "" {
		defaultType = "app"
	}
	fmt.Printf("Type (app/module) [%s]: ", defaultType)
	typ, _ := reader.ReadString('\n')
	typ = strings.TrimSpace(typ)
	if typ == "" {
		typ = defaultType
	}
	entry.Type = typ

	// Try to auto-detect GitHub remote if not in config
	repoURL := defaultRepo
	if repoURL == "" {
		repoURL = getGitRemoteURL()
	}
	owner := ""
	if repoURL != "" {
		parts := strings.Split(repoURL, "/")
		if len(parts) >= 2 {
			owner = parts[len(parts)-2]
		}
	}
	fmt.Printf("Owner (GitHub username/org) [%s]: ", owner)
	ownerInput, _ := reader.ReadString('\n')
	ownerInput = strings.TrimSpace(ownerInput)
	if ownerInput != "" {
		owner = ownerInput
	}
	entry.Owner = owner

	fmt.Print("Description: ")
	entry.Description, _ = reader.ReadString('\n')
	entry.Description = strings.TrimSpace(entry.Description)

	defaultModulesStr := strings.Join(defaultModules, ",")
	fmt.Printf("Comma-separated DEX modules used [%s]: ", defaultModulesStr)
	modulesInput, _ := reader.ReadString('\n')
	modulesInput = strings.TrimSpace(modulesInput)
	if modulesInput == "" {
		entry.DexModules = defaultModules
	} else {
		entry.DexModules = splitAndTrim(modulesInput)
	}

	fmt.Printf("Repo URL [%s]: ", repoURL)
	repoInput, _ := reader.ReadString('\n')
	repoInput = strings.TrimSpace(repoInput)
	if repoInput != "" {
		repoURL = repoInput
	}
	entry.Repo = repoURL

	data, _ := json.MarshalIndent(entry, "", "  ")
	_ = os.WriteFile(".portfolidex.json", data, 0644)
	fmt.Println("Created .portfolidex.json")
}

func splitAndTrim(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func getGitRemoteURL() string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func readProgramidexConfig() *ProgramidexConfig {
	path := ".programidex/init_config.json"
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var cfg ProgramidexConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil
	}
	return &cfg
}

func getDexModulesFromConfig(cfg *ProgramidexConfig) []string {
	var modules []string
	for _, dir := range cfg.Directories {
		if strings.HasPrefix(dir, "modules/") {
			parts := strings.Split(dir, "/")
			if len(parts) > 1 && parts[1] != "" {
				modules = append(modules, parts[1])
			}
		}
	}
	return modules
}

const (
	dexDir                = ".dex"
	programidexConfigFile = ".programidex.json"
	dexLogFile            = "dex.log"
	portfolidexConfigFile = ".portfolidex.json"
)

func init() {
	os.MkdirAll(dexDir, 0755)
	//configPath := filepath.Join(dexDir, portfolidexConfigFile)
	//logPath := filepath.Join(dexDir, dexLogFile)
	//programidexConfigPath := filepath.Join(dexDir, programidexConfigFile)
}

func appendLog(appName, msg string) {
	logPath := filepath.Join(dexDir, dexLogFile)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		timestamp := time.Now().Format(time.RFC3339)
		f.WriteString(fmt.Sprintf("[%s][%s] %s\n", appName, timestamp, msg))
	}
}
