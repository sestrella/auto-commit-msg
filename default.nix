{ pkgs }:

pkgs.buildGoApplication {
  pname = "autocommitmsg";
  version = pkgs.lib.trim (builtins.readFile ./version.txt);
  src = ./.;
  modules = ./gomod2nix.toml;
  meta.mainProgram = "autocommitmsg";
}
