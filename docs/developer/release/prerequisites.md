# Prerequisites

Before performing a release, make sure that you have completed the following.

## Steps

1. Install [helm-docs](https://github.com/norwoodj/helm-docs) on macOS/Linux. This will be used when updating Helm charts.

2. Follow the GitHub [instructions](https://docs.github.com/en/authentication/managing-commit-signature-verification) to set up GPG for signature verification.

3. Optional: Configure git to always sign on commit or tag.

```bash
git config --global commit.gpgSign true
git config --global tag.gpgSign true
```

If you are on macOS or linux and using an encrypted GPG key, `gpg-agent` or `gpg` may be unable
to prompt you for your private key passphrase. This will be denoted by an error
when creating a commit or tag. To circumvent the error, add the following into
your `~/.bash_profile`, `~/.bashrc` or `~/.zshrc`, depending on which shell you are using.

```
export GPG_TTY=$(tty)
```