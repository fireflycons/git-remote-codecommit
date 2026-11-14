// Command git-remote-codecommit is a git remote helper for AWS CodeCommit.
// Git invokes it as: git-remote-codecommit <remote-name> <url>
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/fireflycons/git-remote-codecommit/internal/remoteurl"
	"github.com/fireflycons/git-remote-codecommit/internal/signer"
)

func main() {
	if len(os.Args) < 3 {
		fail("Too few arguments. This hook requires the git command and remote.")
	}
	if len(os.Args) > 3 {
		fail(fmt.Sprintf("Too many arguments. Hook only accepts the git command and remote, but argv was: '%s'", strings.Join(os.Args, " ")))
	}

	signedURL, err := authenticatedURL(context.Background(), os.Args[2])
	if err != nil {
		fail(err.Error())
	}

	// stdout is git's protocol channel; the child inherits all three streams.
	// Credential helpers are disabled: the signed credentials are in the URL, are
	// short-lived, and with a session token exceed some stores' size limits.
	cmd := exec.Command("git", "-c", "credential.helper=", "remote-http", os.Args[1], signedURL)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		fail(err.Error())
	}
}

func authenticatedURL(ctx context.Context, rawURL string) (string, error) {
	remote, err := remoteurl.Parse(rawURL)
	if err != nil {
		return "", err
	}

	profileName := remote.Profile
	var opts []func(*config.LoadOptions) error
	if remote.Profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(remote.Profile))
	} else {
		profileName = "default"
	}
	if remote.Region != "" {
		opts = append(opts, config.WithRegion(remote.Region))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		var notExist config.SharedConfigProfileNotExistError
		if errors.As(err, &notExist) {
			return "", fmt.Errorf("The following profile was not found: %s. Check the profile name in the URL and your AWS config files.", remote.Profile)
		}
		return "", err
	}

	if cfg.Region == "" {
		return "", fmt.Errorf("The following profile does not have an AWS Region: %s. You must set an AWS Region for this profile, or use codecommit::<region>://<profile>@<repository>.", profileName)
	}

	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		var mfa config.AssumeRoleTokenProviderNotSetError
		if errors.As(err, &mfa) {
			return "", fmt.Errorf("Profile %s requires an MFA token, which this helper cannot prompt for. Run git under 'aws-vault exec' (with a URL that has no profile), or use a credential_process profile.", profileName)
		}
		return "", fmt.Errorf("The following profile does not have credentials configured: %s. %v", profileName, err)
	}

	return signer.GitURL(remote.Repository, cfg.Region, signer.Credentials{
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		SessionToken:    creds.SessionToken,
	}, time.Now()), nil
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
