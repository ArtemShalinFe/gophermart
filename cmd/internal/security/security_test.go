package security

import "testing"

func Test_hash_CheckPasswordHash(t *testing.T) {
	hashc, err := NewHashController()
	if err != nil {
		t.Error(err)
	}

	h, err := hashc.HashPassword("test")
	if err != nil {
		t.Error(err)
	}

	type args struct {
		hash     string
		password string
	}
	tests := []struct {
		name string
		h    *hash
		args args
		want bool
	}{
		{
			name: "Positive case",
			h:    hashc,
			args: args{
				hash:     h,
				password: "test",
			},
			want: true,
		},
		{
			name: "Negative case",
			h:    hashc,
			args: args{
				hash:     h,
				password: "tst",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			h := &hash{}
			if got := h.CheckPasswordHash(tt.args.hash, tt.args.password); got != tt.want {
				t.Errorf("hash.CheckPasswordHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
