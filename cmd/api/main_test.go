package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrescan(t *testing.T) {
	tests := []struct {
		name   string
		env    string
		args   []string
		want   string
	}{
		{name: "empty", want: ""},
		{name: "env var", env: "/from/env.yaml", want: "/from/env.yaml"},
		{name: "long space-separated", args: []string{"--config", "/cfg.yaml", "api"}, want: "/cfg.yaml"},
		{name: "long equals", args: []string{"--config=/cfg.yaml", "api"}, want: "/cfg.yaml"},
		{name: "short space-separated", args: []string{"-c", "/cfg.yaml", "api"}, want: "/cfg.yaml"},
		{name: "short equals", args: []string{"-c=/cfg.yaml", "api"}, want: "/cfg.yaml"},
		{name: "env takes precedence over flag", env: "/env.yaml", args: []string{"--config", "/flag.yaml"}, want: "/env.yaml"},
		{name: "config flag at end without value", args: []string{"api", "--config"}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, prescan(tt.env, tt.args))
		})
	}
}
