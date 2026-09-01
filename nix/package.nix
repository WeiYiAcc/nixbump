# nixbump release binary for home-manager (HM-SW)
#
# 分工: GitHub Actions 编译多平台二进制 -> GitHub Release 资产
#       这里 fetchurl 下载 + 已验证 sha256 校验 (不构建)
# 来源: release v0.1.0 (tag aac7d760), hash 实测 2026-08-30
# 用法: home.packages = [ (pkgs.callPackage ./nixbump.nix {}) ];

{ lib, stdenv, fetchurl }:

let
  version = "0.1.0";
  assets = {
    x86_64-linux = {
      url = "https://github.com/WeiYiAcc/nixbump/releases/download/v${version}/nixbump-linux-amd64.tar.gz";
      hash = "sha256-24o3JGZXaYMoOP0seuDzA2kvjD/D5Rez9ksocEGuHnY=";
    };
    aarch64-linux = {
      url = "https://github.com/WeiYiAcc/nixbump/releases/download/v${version}/nixbump-linux-arm64.tar.gz";
      hash = "sha256-UO3An3138T5a9pWh/bhvxwfBmku2GkD6/1VfKKyDNlc=";
    };
    x86_64-darwin = {
      url = "https://github.com/WeiYiAcc/nixbump/releases/download/v${version}/nixbump-darwin-amd64.tar.gz";
      hash = "sha256-BZ7x6al4QZ+pXy0X7Bbqq9dN3NeS2SsY+VzD7xafowM=";
    };
    aarch64-darwin = {
      url = "https://github.com/WeiYiAcc/nixbump/releases/download/v${version}/nixbump-darwin-arm64.tar.gz";
      hash = "sha256-Wejy/Uhv2gf1e5rrO4yyWvVQrXBNNB8hcdKFCX9uLsQ=";
    };
  };
  a = assets.${stdenv.hostPlatform.system}
    or (throw "nixbump: unsupported system ${stdenv.hostPlatform.system}");
in
stdenv.mkDerivation {
  pname = "nixbump";
  inherit version;

  src = fetchurl { inherit (a) url hash; };

  dontUnpack = true;

  installPhase = ''
    install -Dm755 $src $out/bin/nixbump
  '';

  meta = with lib; {
    description = "nixbump: auto-discover and update custom Nix packages (npm/GitHub/GitLab)";
    homepage = "https://github.com/WeiYiAcc/nixbump";
    license = licenses.mit;
    platforms = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
    mainProgram = "nixbump";
  };
}
