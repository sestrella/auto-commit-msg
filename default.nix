{ pkgs }:

pkgs.buildGoApplication {
  pname = "autocommitmsg";
  version = "0.1.0";
  src = ./.;
  modules = ./gomod2nix.toml;
  meta.mainProgram = "autocommitmsg";
}
