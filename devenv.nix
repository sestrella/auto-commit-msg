{ pkgs, ... }:

{
  dotenv.enable = true;

  packages = [
    pkgs.cobra-cli
    pkgs.gitleaks
    pkgs.gomod2nix
  ];

  languages.go.enable = true;

  # See full reference at https://devenv.sh/reference/options/
}
