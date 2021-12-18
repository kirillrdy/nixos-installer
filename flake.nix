{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, utils }:
    let out = system:
      let pkgs = nixpkgs.legacyPackages."${system}";
      in
      {

        devShell = pkgs.mkShell {
          buildInputs = with pkgs; [
            python3Packages.poetry
          ];
        };

        defaultPackage = pkgs.stdenv.mkDerivation {
          name = "nixos-installer";
          src = ./.;
          buildPhase = ''
            HOME=$TMPDIR
            ${pkgs.go}/bin/go build -o nixos-installer main.go

          '';
          installPhase = ''
            mkdir -p $out/bin
            mv nixos-installer $out/bin
          '';
        };

        defaultApp = utils.lib.mkApp {
          drv = self.defaultPackage."${system}";
        };

      }; in with utils.lib; eachSystem defaultSystems out;

}
