package config

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
)

type Config struct {
	// stores a VM's IP address as a string
	VMIP string `json:"vm_ip"`
}
// return the path to the config file 
func getConfigPath() string {
	usr, err := user.Current()
	if err != nil {
		return filepath.Join(".stocknet", "config.json")
	}
	return filepath.Join(usr.HomeDir, ".stocknet", "config.json")
}

var configPath = getConfigPath()

// LoadConfig loads the configuration from file
func LoadConfig() (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil // Return empty config if file doesn't exist
		}
		return nil, err
	}
	// parse JSON and fill config struct with values
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SaveConfig saves the configuration to file
func SaveConfig(config *Config) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// GetVMIP returns the stored VM IP
func GetVMIP() (string, error) {
	config, err := LoadConfig()
	if err != nil {
		return "", err
	}
	return config.VMIP, nil
}

// SetVMIP saves the VM IP
func SetVMIP(ip string) error {
	config := &Config{VMIP: ip}
	return SaveConfig(config)
}
