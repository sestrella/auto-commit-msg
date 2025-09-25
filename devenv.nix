{ pkgs, ... }:

{
  dotenv.enable = true;

  packages = [
    pkgs.cobra-cli
    pkgs.gitleaks
    pkgs.gomod2nix
  ];

  languages.go.enable = true;

  git-hooks.hooks.autocommitmsg =
    let
      autocommitmsg = pkgs.callPackage ./default.nix { };
    in
    {
      enable = true;
      entry = "${autocommitmsg}/bin/autocommitmsg";
      stages = [ "prepare-commit-msg" ];
    };

  # See full reference at https://devenv.sh/reference/options/
}
