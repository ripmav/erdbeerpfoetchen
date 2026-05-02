package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ripmav/erdbeerpfoetchen/internal/config"
)

func writeYAML(t *testing.T, content string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(f, []byte(content), 0600))
	return f
}

func TestNew_ParsesNestedKeys(t *testing.T) {
	path := writeYAML(t, `
debug: true
server:
  listen: ":9090"
  unsafe: true
  tls:
    cert: /tmp/cert.pem
    key: /tmp/key.pem
database:
  uri: postgres://user:pass@localhost/db
`)

	resolver, err := config.New(path)
	require.NoError(t, err)

	var cfg struct {
		Debug         bool   `name:"debug"`
		ServerListen  string `name:"server.listen"`
		ServerUnsafe  bool   `name:"server.unsafe"`
		TLSCert       string `name:"server.tls.cert"`
		TLSKey        string `name:"server.tls.key"`
		DatabaseURI   string `name:"database.uri"`
	}

	k := kong.Must(&cfg, kong.Resolvers(resolver))
	_, err = k.Parse(nil)
	require.NoError(t, err)

	assert.True(t, cfg.Debug)
	assert.Equal(t, ":9090", cfg.ServerListen)
	assert.True(t, cfg.ServerUnsafe)
	assert.Equal(t, "/tmp/cert.pem", cfg.TLSCert)
	assert.Equal(t, "/tmp/key.pem", cfg.TLSKey)
	assert.Equal(t, "postgres://user:pass@localhost/db", cfg.DatabaseURI)
}

func TestNew_CLIOverridesYAML(t *testing.T) {
	path := writeYAML(t, `
server:
  listen: ":9090"
`)

	resolver, err := config.New(path)
	require.NoError(t, err)

	var cfg struct {
		ServerListen string `name:"server.listen" default:":8080"`
	}

	k := kong.Must(&cfg, kong.Resolvers(resolver))
	_, err = k.Parse([]string{"--server.listen=:7070"})
	require.NoError(t, err)

	assert.Equal(t, ":7070", cfg.ServerListen)
}

func TestNew_MissingFile(t *testing.T) {
	_, err := config.New("/nonexistent/path/config.yaml")
	assert.ErrorContains(t, err, "read config file")
}

func TestNew_InvalidYAML(t *testing.T) {
	path := writeYAML(t, ":\tinvalid: yaml: [")
	_, err := config.New(path)
	assert.ErrorContains(t, err, "parse config file")
}

func TestNew_EnvVarOverridesYAML(t *testing.T) {
	path := writeYAML(t, `
server:
  listen: ":9090"
`)

	resolver, err := config.New(path)
	require.NoError(t, err)

	t.Setenv("SERVER_LISTEN", ":7070")

	var cfg struct {
		ServerListen string `name:"server.listen" env:"SERVER_LISTEN" default:":8080"`
	}

	k := kong.Must(&cfg, kong.Resolvers(resolver))
	_, err = k.Parse(nil)
	require.NoError(t, err)

	assert.Equal(t, ":7070", cfg.ServerListen, "env var must override YAML value")
}

func TestNew_SequenceRejected(t *testing.T) {
	path := writeYAML(t, `
server:
  allowed_origins:
    - https://example.com
`)
	_, err := config.New(path)
	assert.ErrorContains(t, err, "YAML sequences are not supported")
}

func TestNew_YAMLListAtRoot(t *testing.T) {
	path := writeYAML(t, "- foo\n- bar\n")
	_, err := config.New(path)
	assert.ErrorContains(t, err, "parse config file")
}

func TestNew_NullValuesSkipped(t *testing.T) {
	path := writeYAML(t, `
server:
  listen: ~
`)

	resolver, err := config.New(path)
	require.NoError(t, err)

	var cfg struct {
		ServerListen string `name:"server.listen" default:":8080"`
	}

	k := kong.Must(&cfg, kong.Resolvers(resolver))
	_, err = k.Parse(nil)
	require.NoError(t, err)

	assert.Equal(t, ":8080", cfg.ServerListen, "null YAML value should leave flag at its default")
}
