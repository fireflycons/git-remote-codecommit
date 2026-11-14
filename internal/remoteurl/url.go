// Package remoteurl parses the URL git passes to git-remote-codecommit.
package remoteurl

import (
	"fmt"
	"regexp"
	"strings"
)

// Remote is the parsed form of a codecommit remote URL.
type Remote struct {
	Profile    string // empty when the URL has no "profile@" part
	Region     string // empty when the region must come from AWS config
	Repository string
}

var regionPattern = regexp.MustCompile(`^[a-z]{2}(-[a-z]+)+-\d$`)

// Parse accepts the two forms git hands to the helper:
//
//	codecommit://[profile@]repository
//	<region>://[profile@]repository   (git strips the "codecommit::" prefix)
func Parse(raw string) (Remote, error) {
	scheme, rest, ok := strings.Cut(raw, "://")
	if !ok {
		return Remote{}, formatError(raw)
	}

	var r Remote
	if scheme != "codecommit" {
		if !regionPattern.MatchString(scheme) {
			return Remote{}, formatError(raw)
		}
		r.Region = scheme
	}

	if profile, repo, found := strings.Cut(rest, "@"); found {
		if profile == "" {
			return Remote{}, formatError(raw)
		}
		r.Profile, rest = profile, repo
	}

	if rest == "" || strings.ContainsAny(rest, "/@") {
		return Remote{}, formatError(raw)
	}
	r.Repository = rest
	return r, nil
}

func formatError(raw string) error {
	return fmt.Errorf("The following URL is malformed: %s. A URL must be in one of the two following formats: codecommit://<profile>@<repository> or codecommit::<region>://<profile>@<repository>", raw)
}
