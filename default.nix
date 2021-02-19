{ buildGoModule
, nix-gitignore
}:

buildGoModule {
  pname = "archetype-go-axon";
  version = "0.0.1";
  src = nix-gitignore.gitignoreSource [] ./.;
  goPackagePath = "github.com/dendrite2go/archetype-go-axon";
  goDeps = ./deps.nix;
  modSha256 = "0052p5n1mnpiklgbp13720s0n465vv4bp4xmyjkrr2lgl130xlld";
}
