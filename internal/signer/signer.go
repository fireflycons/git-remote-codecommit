// Package signer builds the SigV4-authenticated HTTPS URL for a CodeCommit repository,
// matching the algorithm used by aws/git-remote-codecommit.
package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"
)

// Credentials are the AWS credentials used for signing.
type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

// Hostname returns the CodeCommit git endpoint for region, or $CODE_COMMIT_ENDPOINT if set.
func Hostname(region string) string {
	if h := os.Getenv("CODE_COMMIT_ENDPOINT"); h != "" {
		return h
	}
	domain := "amazonaws.com"
	if region == "cn-north-1" || region == "cn-northwest-1" {
		domain = "amazonaws.com.cn"
	}
	return fmt.Sprintf("git-codecommit.%s.%s", region, domain)
}

// GitURL returns https://<username>:<password>@<host>/v1/repos/<repository>.
func GitURL(repository, region string, creds Credentials, now time.Time) string {
	host := Hostname(region)
	path := "/v1/repos/" + repository

	username := creds.AccessKeyID
	if creds.SessionToken != "" {
		username += "%" + creds.SessionToken
	}

	return fmt.Sprintf("https://%s:%s@%s%s", quote(username), password(host, path, region, creds.SecretAccessKey, now), host, path)
}

func password(host, path, region, secret string, now time.Time) string {
	ts := now.UTC().Format("20060102T150405")
	date := ts[:8]

	canonical := "GIT\n" + path + "\n\nhost:" + host + "\n\nhost\n"
	canonicalHash := sha256.Sum256([]byte(canonical))
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		ts,
		date + "/" + region + "/codecommit/aws4_request",
		hex.EncodeToString(canonicalHash[:]),
	}, "\n")

	key := hmacSHA256([]byte("AWS4"+secret), date)
	key = hmacSHA256(key, region)
	key = hmacSHA256(key, "codecommit")
	key = hmacSHA256(key, "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(key, stringToSign))

	return ts + "Z" + signature
}

func hmacSHA256(key []byte, msg string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(msg))
	return h.Sum(nil)
}

// quote percent-encodes everything except RFC 3986 unreserved characters,
// matching Python's urllib.parse.quote(s, safe='').
func quote(s string) string {
	const hexDigits = "0123456789ABCDEF"
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' || c == '-' || c == '_' || c == '.' || c == '~' {
			b.WriteByte(c)
		} else {
			b.WriteByte('%')
			b.WriteByte(hexDigits[c>>4])
			b.WriteByte(hexDigits[c&0xF])
		}
	}
	return b.String()
}
