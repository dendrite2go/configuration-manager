{ buildGoModule
, nix-gitignore
}:

buildGoModule {
  pname = "archetype-go-axon";
  version = "0.0.1";
  src = nix-gitignore.gitignoreSource [] ./.;
  goPackagePath = "github.com/dendrite2go/archetype-go-axon";
  goDeps = ./deps.nix;
  modSha256 = "0xnbpgzw5rs79hns2fwj7sqld12fdlcqswfsgn0dvhlwg7khnp00";
}
