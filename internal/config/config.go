package config

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"gopkg.in/yaml.v3"
)

// Resolver is a kong.Resolver that supplies flag values from a YAML file.
// Precedence: CLI flags > env vars > YAML file > flag defaults.
type Resolver struct {
	values map[string]string
}

// New reads the YAML file at path and returns a Resolver ready for use with kong.Resolvers.
func New(path string) (*Resolver, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	values := make(map[string]string)
	if err := flatten(raw, "", values); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return &Resolver{values: values}, nil
}

// Validate satisfies kong.Resolver. No structural validation is needed because
// unknown keys in the YAML file are simply ignored by the resolver.
func (r *Resolver) Validate(_ *kong.Application) error { return nil }

func (r *Resolver) Resolve(_ *kong.Context, _ *kong.Path, flag *kong.Flag) (interface{}, error) {
	// Env vars take precedence over the YAML file. Kong only checks env vars
	// when the resolver returns nil, so we yield here if any env var for this
	// flag is already set.
	for _, env := range flag.Tag.Envs {
		if _, ok := os.LookupEnv(env); ok {
			return nil, nil
		}
	}

	v, ok := r.values[flag.Name]
	if !ok {
		return nil, nil
	}
	return v, nil
}

// flatten recursively walks a YAML map and writes dot-joined keys into out.
// Example: {server: {listen: ":8080"}} → {"server.listen": ":8080"}
// YAML sequences are rejected because kong has no list-valued flags.
func flatten(m map[string]interface{}, prefix string, out map[string]string) error {
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case nil:
			// skip explicit nulls — let the flag keep its default
		case map[string]interface{}:
			if err := flatten(val, key, out); err != nil {
				return err
			}
		case []interface{}:
			return fmt.Errorf("key %q: YAML sequences are not supported", key)
		default:
			out[key] = fmt.Sprintf("%v", val)
		}
	}
	return nil
}
