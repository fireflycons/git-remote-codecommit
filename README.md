# git-remote-codecommit (Go)

A Go port of [aws/git-remote-codecommit](https://github.com/aws/git-remote-codecommit): a git remote helper that lets you clone, pull, and push AWS CodeCommit repositories using AWS credentials. You don't need CodeCommit HTTPS git credentials or SSH keys. It is a single static binary with no Python dependency.

## Install

```sh
go install github.com/fireflycons/git-remote-codecommit@latest
```

Make sure `git-remote-codecommit` (`.exe` on Windows) is on your `PATH`.

## Usage

```sh
git clone codecommit://MyRepo                        # default profile, its region
git clone codecommit://my-profile@MyRepo             # named profile, its region
git clone codecommit::eu-west-1://my-profile@MyRepo  # named profile, explicit region
git clone codecommit::eu-west-1://MyRepo             # default credential chain, explicit region
```

Credentials and region are resolved by the AWS SDK for Go v2. That covers environment variables, shared config/credentials files, SSO, `credential_process`, and assume-role.

`CODE_COMMIT_ENDPOINT` overrides the git endpoint hostname, for example to use a VPC endpoint.

## MFA-protected roles

The helper cannot prompt for an MFA code, because its stdin carries git's protocol. Use one of these instead:

1. **aws-vault / awsume.** Run git with temporary credentials in the environment, and use a URL **without** a `profile@` part. An explicit profile makes the SDK read that profile instead of the environment.

   ```sh
   aws-vault exec my-role -- git push
   ```

2. **`credential_process`.** Define a profile that gets credentials from aws-vault, and use a GUI MFA prompt. The SDK passes the helper's stdin to the credential process, so a terminal prompt would conflict with git.

   ```ini
   [profile cc]
   region = eu-west-1
   credential_process = aws-vault export --format=json --prompt=wincredui my-role
   ```

   Use `--prompt=osascript` on macOS or `--prompt=zenity` / `kdialog` on Linux. Then `git clone codecommit://cc@MyRepo` works without any wrapper.

## Differences from the Python version

- Regions are checked only for format (`xx-name-N`), not against an endpoint list. An unsupported region fails when git connects.
- Repository names containing `/` or `@` are rejected as malformed.
- Assume-role credentials are not cached on disk between git invocations.
- Git credential helpers are turned off for the HTTPS transport. The signed credentials are short-lived, and with a session token they exceed the Windows Credential Manager's size limit, which made Git Credential Manager print `fatal: Failed to write item to store`.

## Releasing

The version lives in [`VERSION`](VERSION). Every PR into `main` should bump it:

```sh
./scripts/bump-version.sh patch   # or minor / major
```

- Pushing to any branch runs the tests on Linux, macOS and Windows.
- A PR into `main` fails the `version-check` job if tag `v<VERSION>` already exists.
- Merging a PR into `main` creates tag `v<VERSION>` and a GitHub release with binaries for Linux, macOS and Windows (amd64 and arm64), plus `SHA256SUMS`.
- If the PR title contains `[no release]`, the version check and the release are skipped. Use this for docs or CI-only changes.

For the check to actually block merges, add a branch protection rule or ruleset on `main` that requires the `version-check` and `test` status checks.

## License

Apache License 2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE). This is an independent port and is not affiliated with or endorsed by AWS.
