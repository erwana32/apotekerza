package config

import "testing"

func TestLoad_DefaultsAndOverrides(t *testing.T) {
	tests := []struct {
		name       string
		env        map[string]string
		wantAppEnv string
		wantLevel  string
		wantAddr   string
		wantErr    bool
	}{
		{
			name:       "defaults with required database url",
			env:        map[string]string{"DATABASE_URL": "postgres://localhost/apotekerza"},
			wantAppEnv: "local",
			wantLevel:  "info",
			wantAddr:   ":8080",
		},
		{
			name: "environment overrides",
			env: map[string]string{
				"DATABASE_URL": "postgres://localhost/apotekerza",
				"APP_ENV":      "production",
				"LOG_LEVEL":    "debug",
				"HTTP_ADDR":    ":9000",
			},
			wantAppEnv: "production",
			wantLevel:  "debug",
			wantAddr:   ":9000",
		},
		{
			name:    "missing database url",
			env:     map[string]string{},
			wantErr: true,
		},
	}

	keys := []string{"DATABASE_URL", "APP_ENV", "APP_PUBLIC_URL", "LOG_LEVEL", "HTTP_ADDR"}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, key := range keys {
				t.Setenv(key, tt.env[key])
			}

			cfg, err := Load()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if cfg.AppEnv != tt.wantAppEnv || cfg.LogLevel != tt.wantLevel || cfg.HTTPAddr != tt.wantAddr {
				t.Fatalf("Load() = %+v, want env=%s level=%s addr=%s", cfg, tt.wantAppEnv, tt.wantLevel, tt.wantAddr)
			}
		})
	}
}
