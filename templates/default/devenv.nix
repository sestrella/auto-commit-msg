{ pkgs, lib, ... }:

{
  dotenv.enable = true;

  git-hooks.hooks.auto-commit-msg = {
    enable = true;
    entry = lib.getExe pkgs.autocommitmsg;
    stages = [ "prepare-commit-msg" ];
  };
}
