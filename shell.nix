{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go_1_24
    pkgs.protobuf
    pkgs.openssl
    pkgs.git
  ];

  shellHook = ''
    echo "Welcome :)"
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    export PATH=$PATH:$(go env GOPATH)/bin
  '';
}