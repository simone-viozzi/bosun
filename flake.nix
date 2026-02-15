{
  description = "bosun-go-dev";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          name = "bosun-go-dev";

          buildInputs = with pkgs; [
            # Go toolchain + LSP + debug
            go
            gopls
            delve
            gotools
            gofumpt
            uv
            pre-commit

            docker
            docker-compose
          ];

          uv = pkgs.uv;
          pre-commit = pkgs.pre-commit;

          shellHook = ''
            set -euo pipefail

            # Go workspace
            export GOPATH="$PWD/.direnv/gopath"
            export GOBIN="$GOPATH/bin"
            export PATH="$PWD/bin:$GOBIN:$PATH"

            # symlink dlv in .direnv to make it available to vscode
            mkdir -p $PWD/.direnv/go-tools
            if [ -d "$PWD/.direnv/go-tools/delve" ]; then
              rm -rf $PWD/.direnv/go-tools/delve
            fi
            ln -sf "${pkgs.delve}/bin" $PWD/.direnv/go-tools/delve

            # Reproducible-ish builds and static-ish binaries by default
            export CGO_ENABLED=0
            export GOFLAGS="-trimpath"

            echo "Bosun Go dev shell ready."
          '';
        };
      }
    );
}
