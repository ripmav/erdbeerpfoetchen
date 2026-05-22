package cli

import (
	"os"

	"gopkg.in/yaml.v3"
)

// adminConfig is a minimal YAML struct used only to extract twitch.admin-ids.
// It is intentionally separate from Kong so the field is never reachable via
// CLI flags or environment variables.
type adminConfig struct {
	Twitch struct {
		AdminIDs []string `yaml:"admin-ids"`
	} `yaml:"twitch"`
}

// AdminIDs returns the Twitch user IDs designated as admins from the config
// file. Returns nil when no config file is set or the key is absent.
func (cfg *Config) AdminIDs() ([]string, error) {
	if cfg.ConfigFile == "" {
		return nil, nil
	}
	data, err := os.ReadFile(cfg.ConfigFile)
	if err != nil {
		return nil, err
	}
	var ac adminConfig
	if err := yaml.Unmarshal(data, &ac); err != nil {
		return nil, err
	}
	return ac.Twitch.AdminIDs, nil
}
