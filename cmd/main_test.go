package main

import "testing"

func Test_run(t *testing.T) {
	type args struct {
		config string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"test", args{"../config.yaml"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := run(tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
