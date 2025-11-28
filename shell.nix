# shell.nix
{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  packages = with pkgs; [
    go
  ];

  # Optional: set environment variables
  TF_IN_AUTOMATION = "1";

  # Optional: show shell info
  shellHook = ''
    echo "Node.js:   $(go version)"
  '';
}