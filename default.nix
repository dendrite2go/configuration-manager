{ buildGoModule
, nix-gitignore
}:

buildGoModule {
  pname = "archetype-go-axon";
  version = "0.0.1";
  src = nix-gitignore.gitignoreSource [] ./.;
  goPackagePath = "github.com/dendrite2go/archetype-go-axon";
  goDeps = ./deps.nix;
  modSha256 = "0gf7m9ih2ib4mg01myxgivah1ss20zr774d1z729mjxh4k2riiph";
}
