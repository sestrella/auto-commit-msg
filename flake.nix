{
  description = "Generates a commit message from a `git diff` using AI";

  inputs = {
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    nix-filter.url = "github:numtide/nix-filter";
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    systems.url = "github:nix-systems/default";
  };

  outputs =
    {
      gomod2nix,
      nix-filter,
      nixpkgs,
      self,
      systems,
      ...
    }:

    {
      packages = nixpkgs.lib.genAttrs (import systems) (
        system:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [
              gomod2nix.overlays.default
              nix-filter.overlays.default
            ];
          };
          lib = pkgs.lib;
        in
        rec {
          benchmark = pkgs.writeShellScriptBin "benchmark" ''
            ${lib.getExe pkgs.hyperfine} \
              --runs 5 \
              --prepare 'sleep 1' \
              --export-markdown benchmark.md \
              ${lib.getExe go} \
              ${lib.getExe rust}
          '';
          default = go;
          go = pkgs.callPackage ./default.nix { };
          rust = pkgs.callPackage ./rust/default.nix { };
        }
      );

      overlays.default = final: prev: {
        auto-commit-msg = self.packages.${prev.system}.default;
      };

      templates.default = {
        description = "Showcase the devenv integration";
        path = ./templates/default;
      };
    };
}
