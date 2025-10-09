# autocommitmsg

[![Build](https://github.com/sestrella/autocommitmsg/actions/workflows/build.yml/badge.svg)](https://github.com/sestrella/autocommitmsg/actions/workflows/build.yml)

Generates a commit message from a `git diff` using AI.

**Features**

- Supports OpenAI-like providers, with Gemini as the default.
- Improve commit messages by switching to a model based on diff size.

![generated-commit](assets/generated-commit.gif)

> [!NOTE] 
> The [commit messages](https://github.com/sestrella/autocommitmsg/commits/main/)
> for this project were generated using this tool.

## Installation

Download the precompiled binary from the [releases] page that matches your current
system. Unzip the file and place the binary in a location available on your
`PATH` environment variable.

<details>
<summary>Instructions for Nix users</summary>

### devenv

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

</details>

## Configuration

`autocommitmsg` can be configured via a `.autocommitmsg.toml` file in
the project's root directory or the user's home directory. The available
configuration parameters are:

- **`trace`**: When `true`, appends autocommitmsg execution traces to the commit message.
  - **Default**: `false`

  The following metrics are appended to the commit message when trace is enabled:

  - **`model`**: The model used to generate the commit message.
  - **`response_time`**: The time it took to get a response from the AI model.
  - **`execution_time`**: The total time it took for the `autocommitmsg` command to execute.

- **`provider.base_url`**: The base URL of the OpenAI-like provider.
  - **Default**: `https://generativelanguage.googleapis.com/v1beta/openai`
- **`provider.api_key`**: The name of the environment variable that contains the API key.
  - **Default**: `GEMINI_API_KEY`
- **`diff.short_model`**: The model to use for diffs with fewer lines than `diff.threshold`.
  - **Default**: `gemini-2.5-flash-lite`
- **`diff.long_model`**: The model to use for diffs with more lines than `diff.threshold`.
  - **Default**: `gemini-2.5-flash`
- **`diff.threshold`**: The line count threshold to switch between `diff.short_model` and `diff.long_model`.
  - **Default**: `500`

Here is an example `.autocommitmsg.toml` file:

```toml
trace = true

[provider]
base_url = "https://api.openai.com/v1"
api_key = "OPENAI_API_KEY"

[diff]
short_model = "gpt-4.1-mini"
long_model = "o4-mini"
threshold = 250
```

## Usage

After setting `autocommitmsg` as a [prepare-commit-msg] hook, invoking `git
commit` without a commit message generates a commit message. If a commit message
is given, `autocommitmsg` does not generate a commit message and instead uses
the one provided by the user.

![custom-commit](assets/custom-commit.gif)

## License

[MIT](LICENSE)

[prepare-commit-msg]: https://git-scm.com/docs/githooks#_prepare_commit_msg
[releases]: https://github.com/sestrella/autocommitmsg/releases
