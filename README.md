# 🧹 Clean Forks extension for GitHub CLI

Clean Forks is a GitHub CLI extension that helps you clean up your forked repositories.

- ☂️ Dry run mode to see what would be deleted
- 🔑 Seemless authentication with the GitHub CLI
- 🛡️ Protects against deletion of forks with open pull requests

## Installation

To install the extension, run the following command:

```shell
gh extension install nhedger/gh-clean-forks
```

> [!IMPORTANT]
> Your token must have the `delete_repo` scope to delete forks. When deferring the authentication
> to the GitHub CLI, you must ensure that the `delete_repo` scope is included. If you haven't
> already, you can add the scope to your token by running the following command:
> ```
> gh auth refresh -s delete_repo
> ```

## Usage

### Delete forks

This command will delete all forks that **do not have open pull requests**.

```shell
gh clean-forks
```

### Force-delete forks

This command will delete all forks, **including those with open pull requests**.

```shell
gh clean-forks --force
```

### Dry run

Dry-run mode will show you what would be deleted without actually deleting anything.

```shell
gh clean-forks --dry-run
```

## License

This is open source software released under the [MIT License](./LICENSE.md).
