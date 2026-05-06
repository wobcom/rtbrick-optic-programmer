{
  description = "rtbrick optics programmer";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs = { self, nixpkgs, flake-utils }: {
    overlay = (final: prev: let
      pkgs = import nixpkgs {
        inherit (final.stdenv.hostPlatform) system;
      };
    in
      {
      }
    );
  } // (flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs {
        inherit system;
        overlays = [ self.overlay ];
      };
      minimal-dev-pkgs = with pkgs; [
        gopls
        go
      ];
    in {
      devShell = pkgs.mkShell {
        nativeBuildInputs = minimal-dev-pkgs;
      };
    }
  ));
}
