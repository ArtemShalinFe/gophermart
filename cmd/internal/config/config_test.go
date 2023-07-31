package config

import (
	"reflect"
	"testing"
	"time"
)

func TestGetConfig(t *testing.T) {
	defConfig := &Config{
		Address:         "localhost:8078",
		Accrual:         "localhost:8080",
		DSN:             "",
		Key:             []byte("gophermart"),
		AccrualInterval: 2,
		TokenExp:        1 * time.Hour,
	}

	tests := []struct {
		want    *Config
		name    string
		wantErr bool
	}{
		{
			name: "check def values",
			want: defConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
