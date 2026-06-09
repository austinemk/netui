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

  tomlFormat = pkgs.formats.toml { };
in
{
  options.programs.linktui = {
    enable = mkEnableOption "linktui";
    package = mkOption {
      type = with types; nullOr package;
      default = linktui;
      description = "The linktui package to use";
    };

    settings = mkOption {
      type = tomlFormat.type;
      default = { };
      example = literalExpression ''
        {
          window = {
            width = 80;
            height = 28;
          };
        }
      '';
      description = "Settings for linktui";
    };
  };
  config = mkIf cfg.enable {
    home.packages = mkIf (cfg.package != null) [ cfg.package ];

    xdg.configFile."linktui/config.toml" = mkIf (cfg.settings != { }) {
      source = tomlFormat.generate "config.toml" cfg.settings;
    };
  };
}
