package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/ini.v1"
	"argus/src/pkg/downloader"
)

// Manager handles loading and saving configuration
type Manager struct {
	configPath string
	confDir    string
}

// NewManager creates a new config manager with the given config file path and conf.d directory
func NewManager(configPath, confDir string) *Manager {
	return &Manager{
		configPath: configPath,
		confDir:    confDir,
	}
}

// Load loads configuration from the main file and overrides from conf.d directory
func (m *Manager) Load() (*downloader.Config, error) {
	// Start with default config
	cfg := downloader.DefaultConfig()

	// Load main config file if it exists
	if _, err := os.Stat(m.configPath); err == nil {
		if err := m.loadINIFile(m.configPath, cfg); err != nil {
			return nil, fmt.Errorf("failed to load main config file: %w", err)
		}
	}

	// Load override files from conf.d directory in sorted order
	if m.confDir != "" {
		if err := m.loadConfD(cfg); err != nil {
			return nil, fmt.Errorf("failed to load conf.d files: %w", err)
		}
	}

	return cfg, nil
}

// loadConfD loads all .ini files from conf.d directory in sorted order
func (m *Manager) loadConfD(cfg *downloader.Config) error {
	// Check if conf.d directory exists
	if _, err := os.Stat(m.confDir); os.IsNotExist(err) {
		return nil // No conf.d directory, skip
	}

	// Read all files in conf.d
	entries, err := os.ReadDir(m.confDir)
	if err != nil {
		return fmt.Errorf("failed to read conf.d directory: %w", err)
	}

	// Filter and sort .ini files
	var iniFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".ini") {
			iniFiles = append(iniFiles, entry.Name())
		}
	}

	// Sort files by name for sequential override
	sort.Strings(iniFiles)

	// Load each file in order (later files override earlier ones)
	for _, fileName := range iniFiles {
		filePath := filepath.Join(m.confDir, fileName)
		if err := m.loadINIFile(filePath, cfg); err != nil {
			return fmt.Errorf("failed to load %s: %w", fileName, err)
		}
	}

	return nil
}

// loadINIFile loads a single INI file and updates the config
func (m *Manager) loadINIFile(filePath string, cfg *downloader.Config) error {
	iniFile, err := ini.Load(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse INI file: %w", err)
	}

	// Read values from [github] section
	if iniFile.HasSection("github") {
		section := iniFile.Section("github")
		if val := section.Key("repo_owner").String(); val != "" {
			cfg.RepoOwner = val
		}
		if val := section.Key("repo_name").String(); val != "" {
			cfg.RepoName = val
		}
		if val := section.Key("api_url").String(); val != "" {
			cfg.GitHubAPIURL = val
		}
	}

	// Read values from [download] section
	if iniFile.HasSection("download") {
		section := iniFile.Section("download")
		if val := section.Key("folder").String(); val != "" {
			cfg.DownloadFolder = val
		}
		if val := section.Key("package_pattern").String(); val != "" {
			cfg.PackagePattern = val
		}
	}

	return nil
}

// Save saves the configuration to the main INI file
func (m *Manager) Save(cfg *downloader.Config) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create INI file using LoadSources with empty data
	cfgFile := ini.Empty()

	// GitHub section
	githubSection, err := cfgFile.NewSection("github")
	if err != nil {
		return fmt.Errorf("failed to create github section: %w", err)
	}
	githubSection.Key("repo_owner").SetValue(cfg.RepoOwner)
	githubSection.Key("repo_name").SetValue(cfg.RepoName)
	githubSection.Key("api_url").SetValue(cfg.GitHubAPIURL)

	// Download section
	downloadSection, err := cfgFile.NewSection("download")
	if err != nil {
		return fmt.Errorf("failed to create download section: %w", err)
	}
	downloadSection.Key("folder").SetValue(cfg.DownloadFolder)
	downloadSection.Key("package_pattern").SetValue(cfg.PackagePattern)

	// Write to file
	if err := cfgFile.SaveTo(m.configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// CreateDefault creates a default config file
func (m *Manager) CreateDefault() error {
	cfg := downloader.DefaultConfig()
	return m.Save(cfg)
}
