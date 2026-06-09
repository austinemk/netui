self:
{
  lib,
  pkgs,
  config,
  ...
}:
with lib;
let
  cfg = config.programs.linktui;
  linktui = self.packages.${pkgs.stdenv.hostPlatform.system}.default;
in
{
  options.programs.linktui = {
    enable = mkEnableOption "linktui";
    package = mkOption {
      type = types.package;
      default = linktui;
      description = "The linktui package to use";
    };
  };
  config = mkIf cfg.enable {
    networking.networkmanager.enable = true;
    hardware.bluetooth.enable = true;

    environment.systemPackages = [ cfg.package ];
  };
}
