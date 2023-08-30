package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// KUBECONFIG is merged here
// https://pkg.go.dev/k8s.io/client-go/tools/clientcmd#ClientConfigLoadingRules.Load
// It uses https://github.com/imdario/mergo   mergo.Merge()

// Load reads in config file by searching for the first in configFiles
func Load(log *slog.Logger, scheme *runtime.Scheme, conf runtime.Object, configFiles []string) error {
	codecs := serializer.NewCodecFactory(scheme, serializer.EnableStrict)

	// For now we simply pick the first one.  If we wanted to expand this we could use mergo (see above) to merge the files in reverse order.
	for _, filename := range configFiles {
		content, err := os.ReadFile(filename)
		if err != nil {
			log.Debug("Skipping config file", "path", filename, "reason", err)
			continue
		}

		// Regardless of if the bytes are of any external version,
		// it will be read successfully and converted into the internal version
		if err := runtime.DecodeInto(codecs.UniversalDecoder(), content, conf); err != nil {
			return fmt.Errorf("loading configuration: %w", err)
		}

		log.Info("Using config file", "path", filename)
		break
	}

	// if no files are found then the configuration might not be defaulted so we again to be sure.
	scheme.Default(conf)

	return nil
}

// DefaultConfigSearchPath returns the list of locations to look for configuration files
func DefaultConfigSearchPath(parts ...string) []string {
	return []string{
		strings.Join(parts, "-"),
		filepath.Join(xdg.ConfigHome, filepath.Join(parts...)),
		filepath.Join("/", "etc", filepath.Join(parts...)),
	}
	// TODO we should consider searching $XDG_CONFIG_DIRS as well
}

// DefaultConfigPath is the path we would save the configuration to if needed.  In a sense it is the preferred configuration path.
func DefaultConfigPath(parts ...string) string {
	return filepath.Join(xdg.ConfigHome, filepath.Join(parts...))
}
