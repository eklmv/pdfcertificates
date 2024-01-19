{
  description = "Go development environment";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs }:
    let
      goV = 21;
      overlays = [
        (_: prev: {
          go = prev."go_1_${toString goV}";
        })
      ];
      supportedSystems =
        [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forEachSupportedSystem = f:
        nixpkgs.lib.genAttrs supportedSystems
          (system: f { pkgs = import nixpkgs { inherit overlays system; }; });
    in
    {
      devShells = forEachSupportedSystem ({ pkgs }: {
        default =
          pkgs.mkShell {
            NIX_HARDENING_ENABLE = "";
            packages = with pkgs; [ go gotools gopls gnumake go-migrate sqlc go-mockery ];
          };
        test =
          pkgs.mkShell {
            packages = with pkgs; [ go gnumake ];
          };
      });
    };
}
