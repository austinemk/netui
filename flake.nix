{
  description = "Tui based wifi, bluetooth and vpn manager for linux";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    git-hooks.url = "github:cachix/git-hooks.nix";
    go-overlay.url = "github:purpleclay/go-overlay";
  };

  outputs =
    {
      self,
      nixpkgs,
      go-overlay,
      git-hooks,
      ...
    }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
      overlays = [ (import go-overlay) ];
      forAllSystems =
        f:
        nixpkgs.lib.genAttrs systems (
          system:
          f {
            system = system;
            pkgs = import nixpkgs { inherit system overlays; };
          }
        );
      mkLinktui = pkgs: pkgs.callPackage ./nix/package.nix { go = pkgs.go-bin.fromGoMod ./go.mod; };
    in
    {
      packages = forAllSystems (
        {
          pkgs,
          system,
        }:
        {
          default = mkLinktui pkgs;
          linktui = self.packages.${system}.default;
        }
      );

      overlays = {
        default = final: _: {
          linktui = self.packages.${final.stdenv.hostPlatform.system}.default;
        };
        linktui = self.overlays.default;
      };

      formatter = forAllSystems (
        { pkgs, system }:
        let
          config = self.checks.${system}.pre-commit-check.config;
          inherit (config) package configFile;
          script = ''
            ${pkgs.lib.getExe package} run --all-files --config ${configFile}
          '';
        in
        pkgs.writeShellScriptBin "pre-commit-run" script
      );

      checks = forAllSystems (
        { pkgs, system }:
        {
          pre-commit-check = git-hooks.lib.${system}.run {
            src = ./.;
            hooks = {
              nixfmt.enable = true;
              # golines.enable = true; ? Idk what formatter is used or if there are tests
              # gotest.enable = true;
            };

            package = pkgs.prek;
          };
        }
      );

      devShells = forAllSystems (
        { pkgs, system }:
        let
          go = pkgs.go-bin.fromGoMod ./go.mod;
          inherit (self.checks.${system}.pre-commit-check) shellHook enabledPackages;
        in
        {
          default = pkgs.mkShell {
            buildInputs = [ go-overlay.packages.${system}.govendor ] ++ enabledPackages;
            shellHook = shellHook + ''
              export PATH=${go}/bin:$PATH
              govendor
            '';
          };
        }
      );

      nixosModules.default = import ./nix/nixos.nix self;

      homeModules.default = import ./nix/hm.nix self;
    };
}
