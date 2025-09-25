{ pkgs, ... }:

{
  dotenv.enable = true;

  packages = [
    pkgs.cobra-cli
    pkgs.gitleaks
  ];

  languages.go.enable = true;

  # See full reference at https://devenv.sh/reference/options/
}
