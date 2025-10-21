{
  pkgs,
  lib,
  nix-filter,
}:

pkgs.buildGoApplication {
  pname = "auto-commit-msg";
  version = lib.trim (builtins.readFile ./version.txt);
  src = nix-filter {
    root = ./.;
    include = [
      "cmd"
      "internal"
      ./go.mod
      ./go.sum
      ./main.go
      ./version.txt
    ];
  };
  # TODO: change the name of the binary
  modules = ./gomod2nix.toml;
  meta = {
    description = "Generates a commit message from a git diff using AI";
    license = lib.licenses.mit;
    mainProgram = "auto-commit-msg";
  };
}
