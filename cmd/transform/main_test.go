package main

import "testing"

func Test_convertDateMonth(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "2023-08",
			args: args{s: "2023-08"},
			want: "2023-Aug",
		},
		{
			name: "2023-W41",
			args: args{s: "2023-W41"},
			want: "2023-W41",
		},
		{
			name: "2023-01-01",
			args: args{s: "2023-01-01"},
			want: "2023-01-01",
		},
		{
			name: "2023-Q1",
			args: args{s: "2023-Q1"},
			want: "2023-Q1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertDateMonth(tt.args.s); got != tt.want {
				t.Errorf("convertDateMonth() = %v, want %v", got, tt.want)
			}
		})
	}
}
