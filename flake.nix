{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      utils,
    }:
    let
      out =
        system:
        let
          pkgs = nixpkgs.legacyPackages."${system}";
        in
        {
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
        };
    in
    with utils.lib;
    eachSystem defaultSystems out;

}
