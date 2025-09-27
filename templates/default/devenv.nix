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
