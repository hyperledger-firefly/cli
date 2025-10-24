package stacks
import (
	"os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
    "gopkg.in/yaml.v3"

	"github.com/hyperledger/firefly-cli/pkg/types"
)

func TestMockConfigWithAutoReload(t *testing.T) {
    tmpDir := os.TempDir()
    configDir := filepath.Join(tmpDir, "config")

    // Ensure the config directory exists
    err := os.MkdirAll(configDir, 0755)
    assert.NoError(t, err)

    configPath := filepath.Join(configDir, "firefly_core_member1.yml")

    // Write mock config YAML with autoReload true
    configYAML := []byte(`
autoReload: true
`)

    err = os.WriteFile(configPath, configYAML, 0644)
    assert.NoError(t, err)

    // Read the file back in
    content, err := os.ReadFile(configPath)
    assert.NoError(t, err)

    var config types.FireflyConfig
    err = yaml.Unmarshal(content, &config)
    assert.NoError(t, err)

    // Check that AutoReload is true as expected
    assert.True(t, config.Config.AutoReload)
}
