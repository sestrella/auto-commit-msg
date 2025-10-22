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
  modules = ./gomod2nix.toml;
  postInstall = ''
    mv $out/bin/auto-commit-msg $out/bin/acm
  '';
  meta = {
    description = "Generates a commit message from a git diff using AI";
    license = lib.licenses.mit;
    mainProgram = "acm";
  };
}
