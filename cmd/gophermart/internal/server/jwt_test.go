package server

import (
	"strings"
	"testing"
)

func TestNewJWTToken(t *testing.T) {
	superSecretKey := []byte("SuperSecretKey")

	type args struct {
		login     string
		password  string
		secretKey []byte
	}
	tests := []struct {
		name    string
		want    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive case #1",
			args: args{
				secretKey: superSecretKey,
				login:     "test",
				password:  "test",
			},
			want:    "eyJleHAiOjE2ODkyNDE1MTUsIkxvZ2luIjoidGVzdCIsIlBhc3N3b3JkIjoidGVzdCJ9",
			wantErr: false,
		},
		{
			name: "Positive case #2",
			args: args{
				secretKey: superSecretKey,
				login:     "test",
				password:  "SuperSecretPass",
			},
			want:    "eyJleHAiOjE2ODkyNDE3OTEsIkxvZ2luIjoidGVzdCIsIlBhc3N3b3JkIjoiU3VwZXJTZWNyZXRQYXNzIn0",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewJWTToken(tt.args.secretKey, tt.args.login, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewJWTToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.ContainsAny(got, tt.want) {
				t.Errorf("NewJWTToken() = %v not contain %v", got, tt.want)
			}
		})
	}
}

func TestIsAuthorized(t *testing.T) {
	type args struct {
		login     string
		password  string
		secretKey []byte
	}

	key := []byte("TrueKey")
	brokenkey := []byte("BrokenKey")

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Positive case",
			args: args{
				secretKey: key,
				login:     "test",
				password:  "test",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Negative case",
			args: args{
				secretKey: brokenkey,
				login:     "test",
				password:  "test",
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tokenString, err := NewJWTToken(key, tt.args.login, tt.args.password)
			if err != nil {
				t.Error(err)
			}

			got, err := IsAuthorized(tokenString, tt.args.secretKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsAuthorized() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsAuthorized() = %v, want %v", got, tt.want)
			}
		})
	}
}
