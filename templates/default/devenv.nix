{ pkgs, lib, ... }:

{
  dotenv.enable = true;

  git-hooks.hooks.auto-commit-msg = {
    enable = true;
    entry = lib.getExe pkgs.auto-commit-msg;
    stages = [ "prepare-commit-msg" ];
  };

  cachix.pull = [ "sestrella" ];
}
