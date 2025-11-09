{
  nix-filter,
  pkgs,
  rustPlatform,
}:

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

  buildInputs = [ pkgs.openssl ];

  # nativeBuildInputs = [ pkgs.pkg-config ];

  meta.mainProgram = "acm";
}
