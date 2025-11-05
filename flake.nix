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
        in
        {
          "auto-commit-msg/rust" = pkgs.callPackage ./rust/default.nix { };
          auto-commit-msg = pkgs.callPackage ./default.nix { };
          default = self.packages.${system}.auto-commit-msg;
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
