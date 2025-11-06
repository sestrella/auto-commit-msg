{ nix-filter, rustPlatform }:

rustPlatform.buildRustPackage {
  pname = "auto-commit-msg";
  version = "0.4.0-dev";
  src = nix-filter {
    root = ./.;
    include = [
      "src"
      ./Cargo.lock
      ./Cargo.toml
    ];
  };
  cargoLock = {
    lockFile = ./Cargo.lock;
  };
  meta.mainProgram = "acm";
}
