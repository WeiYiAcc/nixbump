{
  lib,
  buildGoModule,
  nix-update,
  nix,
  git,
  gh,
  makeWrapper,
}:
buildGoModule {
  pname = "nixbump";
  version = "0.1.0";

  src = ./.;

  vendorHash = null;

  nativeBuildInputs = [makeWrapper];

  postFixup = ''
    wrapProgram $out/bin/nixbump \
      --prefix PATH : ${
        lib.makeBinPath [
          nix-update
          nix
          git
          gh
        ]
      }
  '';

  meta = {
    description = "Auto-update Nix package versions and hashes";
    mainProgram = "nixbump";
    license = lib.licenses.mit;
  };
}
