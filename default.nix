{ pkgs, lib }:

pkgs.buildGoApplication {
  pname = "autocommitmsg";
  version = lib.trim (builtins.readFile ./version.txt);
  src = ./.;
  modules = ./gomod2nix.toml;
  meta = {
    description = "Generates a commit message from a git diff using AI";
    license = lib.licenses.mit;
    mainProgram = "autocommitmsg";
  };
}
