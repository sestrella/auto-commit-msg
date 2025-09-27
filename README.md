# autocommitmsg

[![Main](https://github.com/sestrella/autocommitmsg/actions/workflows/main.yml/badge.svg)](https://github.com/sestrella/autocommitmsg/actions/workflows/main.yml)

Generates a commit message from a `git diff` using AI.

## Usage

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

Declare the `autocommitmsg` hook as follows:

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

## License

[MIT](LICENSE)
