#!/usr/bin/bash
set -e

BINARY=q
BINDIR="${BINDIR:-$HOME/.local/bin}"
INSTALL_PATH="$BINDIR/$BINARY"

@all() { @build
}

@build() {
    go build -o "$BINARY" .
}

@install() {
    @build
    mkdir -p "$BINDIR"
    install -m 755 "$BINARY" "$INSTALL_PATH"
}

@uninstall() {
    if [[ -f "$INSTALL_PATH" ]]; then
        rm -f "$INSTALL_PATH"
    fi
}

@test() {
    go test ./... "$@"
}

@run() {
    go run . "$@"
}

@clean() {
    rm -f "$BINARY"
    go clean
}

##################################################################
@help() {
    echo "do™️: Do some commands for this project. Like Just, but in bash and self-contained."
    echo
    echo "Available commands:"
    declare -F | grep "^declare -f @" | cut -f 2 -d @ | sed "s|^|\t$0 |"
}
DEFAULT=help
if [[ -z $1 ]]; then
    eval "@$DEFAULT"
else
    declare -F | grep -qx "declare -f @$1" || { echo "No such task: $1"; echo; @help; exit 1; }
    task=@$1; shift
    eval "$task "$@""
fi
