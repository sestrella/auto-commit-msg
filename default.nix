{ pkgs }:

pkgs.buildGoApplication {
  pname = "auto-commit-msg";
  version = "0.1.0";
  src = ./.;
  modules = ./gomod2nix.toml;
  meta.mainProgram = "auto-commit-msg";
}
