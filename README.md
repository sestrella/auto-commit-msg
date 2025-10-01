# autocommitmsg

[![Main](https://github.com/sestrella/autocommitmsg/actions/workflows/main.yml/badge.svg)](https://github.com/sestrella/autocommitmsg/actions/workflows/main.yml)

Generates a commit message from a `git diff` using AI.

**Features**

- Supports OpenAI-like providers.
- Switch model based on the diff.

## Installation

### devenv users

Add the `autocommitmsg` input to the `devenv.yaml` file:

```yml
inputs:
  autocommitmsg:
    url: github:sestrella/autocommitmsg
    overlays: [default]
  nixpkgs:
    url: github:cachix/devenv-nixpkgs/rolling
```

Add the `autocommitmsg` hook to the `devenv.nix` file as follows:

```nix
{ pkgs, lib, ... }:

{
  dotenv.enable = true;

  git-hooks.hooks.autocommitmsg = {
    enable = true;
    entry = lib.getExe pkgs.autocommitmsg;
    stages = [ "prepare-commit-msg" ];
  };

  cachix.pull = [ "sestrella" ];
}
```

**Note:** Enabling `dotenv` is optional if the `OPENAI_API_KEY` environment
variable is available.

## Usage

After setting `autocommitmsg` as a [prepare-commit-msg] hook, invoking `git
commit` without a commit message generates a commit message.

![generated-commit](assets/generated-commit.gif)

If a commit message is given, `autocommitmsg` does not generate a commit message
and instead uses the one provided by the user.

![custom-commit](assets/custom-commit.gif)

## License

[MIT](LICENSE)

[prepare-commit-msg]: https://git-scm.com/docs/githooks#_prepare_commit_msg
