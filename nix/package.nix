{ buildGoApplication, go }:
buildGoApplication {
  inherit go;
  pname = "linktui";
  version = "git";
  src = ../.;
  modules = ../govendor.toml;
}
