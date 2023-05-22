package params

import (
	"strings"

	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
)

func CustomConfigTemplate() string {
	config := serverconfig.DefaultConfigTemplate
	lines := strings.Split(config, "\n")
	// Note you can add more lines here
	return strings.Join(lines, "\n")
}

// CustomAppConfig defines lum's custom application configuration.
type CustomAppConfig struct {
	serverconfig.Config
}
