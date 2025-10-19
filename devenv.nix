{
  config,
  lib,
  pkgs,
  ...
}:

{
  env.GEMINI_API_KEY = config.secretspec.secrets.GEMINI_API_KEY or "";

  packages = [
    pkgs.asciinema
    pkgs.cobra-cli
    pkgs.gitleaks
    pkgs.gomod2nix
    pkgs.secretspec
  ];

  languages.go.enable = true;

  git-hooks.hooks.auto-commit-msg = {
    enable = true;
    entry = lib.getExe (pkgs.callPackage ./default.nix { });
    stages = [ "prepare-commit-msg" ];
    verbose = true;
  };
}
