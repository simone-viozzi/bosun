{ pkgs ? import <nixpkgs> { } }:

pkgs.mkShell {
  name = "bosun-go-dev";

  buildInputs = with pkgs; [
    # Go toolchain + LSP + debug
    go                # Go compiler
    gopls             # Go language server
    delve             # Go debugger
    gotools           # goimports, stringer, etc.
    gofumpt           # stricter gofmt
    uv
    pre-commit


    # Docker + Compose + Colima
    docker
    docker-compose    # so `docker compose` works in your shell
    colima
  ];

  uv = pkgs.uv;
  pre-commit = pkgs.pre-commit;

  # Useful environment and Docker isolation
  shellHook = ''
    set -euo pipefail

    # Go workspace
    export GOPATH="$PWD/.direnv/gopath"
    export GOBIN="$GOPATH/bin"
    export PATH="$GOBIN:$PATH"

    # symlink dlv in .direnv to make it available to vscode
    mkdir -p $PWD/.direnv/go-tools
    if [ -d "$PWD/.direnv/go-tools/delve" ]; then
      rm -rf $PWD/.direnv/go-tools/delve
    fi
    ln -sf "${pkgs.delve}/bin" $PWD/.direnv/go-tools/delve

    # Reproducible-ish builds and static-ish binaries by default
    export CGO_ENABLED=0
    export GOFLAGS="-trimpath"

    # Spin up an isolated Docker daemon via Colima (profile: bosun-go-dev)
    colima start -p bosun-go-dev --cpu 2 --memory 4 --activate

    # Point Docker CLI/SDK to Colima socket
    export DOCKER_HOST="unix://$HOME/.config/colima/bosun-go-dev/docker.sock"

    echo "Bosun Go dev shell ready. DOCKER_HOST=$DOCKER_HOST"
  '';
}
