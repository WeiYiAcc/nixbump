{
  lib,
  stdenvNoCC,
  fetchurl,
}:
stdenvNoCC.mkDerivation rec {
  pname = "sliver";
  version = "1.6.6";

  srcs = [
    (fetchurl {
      url = "https://github.com/BishopFox/sliver/releases/download/v${version}/sliver-server_linux-amd64";
      hash = "sha256-xj9V9zkHm5iDOq6JdF03OgXPh2alAtvymz81O+GhhFI=";
      name = "sliver-server";
    })
    (fetchurl {
      url = "https://github.com/BishopFox/sliver/releases/download/v${version}/sliver-client_linux-amd64";
      hash = "sha256-iwTip0x8fXK/5mb3dvgZzaLNTJcW2spBge0x7rAUY/E=";
      name = "sliver-client";
    })
  ];
}
