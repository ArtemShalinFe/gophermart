package models

import "testing"

func TestOrderDTONumberIsCorrect(t *testing.T) {

	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{
			name:   "#1 49927398716",
			number: "49927398716",
			want:   true,
		},
		{
			name:   "#2 49927398717",
			number: "49927398717",
			want:   false,
		},
		{
			name:   "#3 1234567812345678",
			number: "1234567812345678",
			want:   false,
		},
		{
			name:   "#4 1234567812345670",
			number: "1234567812345670",
			want:   true,
		},
		{
			name:   "#5 4026843483168683",
			number: "4026843483168683",
			want:   true,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			o := &OrderDTO{
				UserID: "1",
				Number: tt.number,
			}
			if got := o.NumberIsCorrect(); got != tt.want {
				t.Errorf("OrderDTO.NumberIsCorrect() = %v, want %v", got, tt.want)
			}
		})
	}
}
