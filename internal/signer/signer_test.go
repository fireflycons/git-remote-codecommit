package signer

import (
	"testing"
	"time"
)

// Expected values were computed by a stdlib-only Python reimplementation of
// git_remote_codecommit.git_url / botocore SigV4Auth (string_to_sign, signature).
func TestGitURL(t *testing.T) {
	t.Setenv("CODE_COMMIT_ENDPOINT", "")
	now := time.Date(2026, 1, 14, 10, 11, 12, 0, time.UTC)
	const secret = "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY"

	tests := []struct {
		name   string
		region string
		creds  Credentials
		want   string
	}{
		{
			name:   "no session token",
			region: "eu-west-1",
			creds:  Credentials{AccessKeyID: "AKIDEXAMPLE", SecretAccessKey: secret},
			want:   "https://AKIDEXAMPLE:20260114T101112Z003fe8640e6c7903f0c715ad9fa3e6538923c0829be33001aa0a47a7802cfbd9@git-codecommit.eu-west-1.amazonaws.com/v1/repos/MyRepo",
		},
		{
			name:   "session token, china region",
			region: "cn-north-1",
			creds:  Credentials{AccessKeyID: "AKIDEXAMPLE", SecretAccessKey: secret, SessionToken: "Tok/en+abc="},
			want:   "https://AKIDEXAMPLE%25Tok%2Fen%2Babc%3D:20260114T101112Z36e458dfee03ba8bd6e8ecd103079a40ab5662f29f931dcccdbee49353e64e45@git-codecommit.cn-north-1.amazonaws.com.cn/v1/repos/MyRepo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GitURL("MyRepo", tt.region, tt.creds, now); got != tt.want {
				t.Errorf("GitURL()\n got  %s\n want %s", got, tt.want)
			}
		})
	}
}

func TestHostnameOverride(t *testing.T) {
	t.Setenv("CODE_COMMIT_ENDPOINT", "vpce-123.git-codecommit.eu-west-1.vpce.amazonaws.com")
	if got := Hostname("eu-west-1"); got != "vpce-123.git-codecommit.eu-west-1.vpce.amazonaws.com" {
		t.Errorf("Hostname() = %s", got)
	}
}

func TestQuote(t *testing.T) {
	if got, want := quote("aZ09-_.~ /+=%@:"), "aZ09-_.~%20%2F%2B%3D%25%40%3A"; got != want {
		t.Errorf("quote() = %s, want %s", got, want)
	}
}
