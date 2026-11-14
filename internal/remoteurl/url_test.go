package remoteurl

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		in      string
		want    Remote
		wantErr bool
	}{
		{in: "codecommit://MyRepo", want: Remote{Repository: "MyRepo"}},
		{in: "codecommit://dev@MyRepo", want: Remote{Profile: "dev", Repository: "MyRepo"}},
		{in: "eu-west-1://MyRepo", want: Remote{Region: "eu-west-1", Repository: "MyRepo"}},
		{in: "us-gov-west-1://dev@MyRepo", want: Remote{Profile: "dev", Region: "us-gov-west-1", Repository: "MyRepo"}},
		{in: "codecommit://", wantErr: true},
		{in: "codecommit://dev@", wantErr: true},
		{in: "codecommit://@MyRepo", wantErr: true},
		{in: "codecommit://a@b@c", wantErr: true},
		{in: "codecommit://MyRepo/extra", wantErr: true},
		{in: "https://MyRepo", wantErr: true},
		{in: "eu-west://MyRepo", wantErr: true},
		{in: "MyRepo", wantErr: true},
	}
	for _, tt := range tests {
		got, err := Parse(tt.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("Parse(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("Parse(%q) = %+v, want %+v", tt.in, got, tt.want)
		}
	}
}
