package arg

import "testing"

func TestValidateCount(t *testing.T) {
	type args struct {
		args          []string
		expLen        int
		missingErrMsg string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				args:          []string{"hi"},
				expLen:        1,
				missingErrMsg: "",
			},
			wantErr: false,
		},
		{
			name: "",
			args: args{
				args:          []string{"hi"},
				expLen:        2,
				missingErrMsg: "",
			},
			wantErr: true,
		},
		{
			name: "",
			args: args{
				args:          nil,
				expLen:        0,
				missingErrMsg: "",
			},
			wantErr: false,
		},
		{
			name: "",
			args: args{
				args:          nil,
				expLen:        2,
				missingErrMsg: "",
			},
			wantErr: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateCount(tt.args.args, tt.args.expLen, tt.args.missingErrMsg); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
