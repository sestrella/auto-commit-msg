{
  description = "A very basic flake";

  inputs = {
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    systems.url = "github:nix-systems/default";
  };

  outputs =
    {
      self,
      gomod2nix,
      nixpkgs,
      systems,
      ...
    }:
    let
      forAllSystems = nixpkgs.lib.genAttrs (import systems);
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [ gomod2nix.overlays.default ];
          };
        in
        {
          default = pkgs.callPackage ./default.nix { };
        }
      );

      overlays.default = final: prev: {
        autocommitmsg = self.packages.${prev.system}.default;
      };

      templates.default = {
        description = "";
        path = ./templates/default;
      };
    };
}
