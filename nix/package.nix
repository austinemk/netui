{
  buildGoApplication,
  lib,
  go,
}:
buildGoApplication {
  inherit go;
  pname = "linktui";
  version = "git";
  src = ../.;
  modules = ../govendor.toml;

  meta = with lib; {
    description = "Tui based wifi, bluetooth and vpn manager for linux";
    homepage = "https://github.com/austinemk/linktui";
    license = licenses.mit;
    maintainers = with maintainers; [ Immelancholy ];
    mainProgram = "linktui";
  };
}
