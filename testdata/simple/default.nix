{
  lib,
  stdenvNoCC,
  fetchurl,
}:
stdenvNoCC.mkDerivation rec {
  pname = "atuin-desktop-bin";
  version = "0.2.20";

  src = fetchurl {
    url = "https://github.com/atuinsh/desktop/releases/download/v${version}/Atuin_Desktop_${version}_amd64.deb";
    hash = "sha256-DTvxZ8CTdp9gRC3rrrIWn8RIn8pfe1IPn3KldmVJ18c=";
  };
}
