{ pkgs, lib, ... }:

{
  dotenv.enable = true;

  packages = [
    pkgs.cobra-cli
    pkgs.gitleaks
    pkgs.gomod2nix
  ];

  languages.go.enable = true;

  git-hooks.hooks.auto-commit-msg = {
    enable = true;
    entry = lib.getExe (pkgs.callPackage ./default.nix { });
    stages = [ "prepare-commit-msg" ];
    verbose = true;
  };
}
