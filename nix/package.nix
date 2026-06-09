{ buildGoApplication, go }:
buildGoApplication {
  inherit go;
  pname = "linktui";
  version = "0.1.0";
  src = ../.;
  modules = ../govendor.toml;
}
