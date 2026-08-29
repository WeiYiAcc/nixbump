# nixbump release binary for home-manager
#
# 分工: GitHub Actions 编译多平台二进制 -> release tarball
#       这里只 fetchurl 下载 + hash 校验 (不构建)
# 用法:
#   home.packages = [ (pkgs.callPackage ./nixbump.nix {}) ];
#
# 版本/hash 更新: nix-prefetch-url 或 nixbump 自身 -check 后填入

{ lib
, stdenv
, fetchurl
, version ? "0.1.0"
}:

let
  system = stdenv.hostPlatform.system;
  goarch = {
    x86_64-linux  = "amd64";
    aarch64-linux = "arm64";
    x86_64-darwin = "amd64";
    aarch64-darwin = "arm64";
  }.${system} or (throw "unsupported system: ${system}");
  goos = if stdenv.hostPlatform.isDarwin then "darwin" else "linux";
in
stdenv.mkDerivation {
  pname = "nixbump";
  inherit version;

  src = fetchurl {
    url = "https://github.com/WeiYiAcc/nixbump/releases/download/v${version}/nixbump-${goos}-${goarch}";
    # 首次 build 后 nix 报错填入真实 sha256:
    #   error: hash mismatch ... got: sha256-xxxx
    sha256 = lib.fakeSha256;
  };

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
