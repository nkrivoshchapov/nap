self:
{
  lib,
  pkgs,
  config,
  ...
}:
let
  inherit (lib)
    mkIf
    types
    mkOption
    mkEnableOption
    mkPackageOption
    ;

  settingsFormat = pkgs.formats.yaml { };
in
{
  meta.maintainers = [ lib.maintainers.isabelroses ];

  options.programs.nap = {
    enable = mkEnableOption "A code snippet manager for your terminal";

    package = mkPackageOption self.packages.${pkgs.stdenv.hostPlatform.system} "nap" { };

    settings = {
      home = mkOption {
        type = with types; str;
        default = "~/.nap";
        example = "~/.snippets";
        description = "home for your snippets";
      };

      default_language = mkOption {
        type = with types; str;
        default = "go";
        example = "rust";
        description = "default language for new snippets";
      };

      theme = mkOption {
        type = with types; str;
        default = "catppuccin-mocha";
        example = "nord";
        description = "theme for code previews";
      };
    };

    colors = mkOption {
      inherit (settingsFormat) type;
      default = { };
      example = lib.literalExpression ''
        background = "#252B2E";
        foreground = "#D9E4DC";
        primary_color = "#B2C98F";
        primary_color_subdued = "#6E8585";
        green = "#B2C98F";
        bright_green = "#83C092";
        red = "#E67E80";
        bright_red = "#E69875";
        textinvert = "#46545B";
        gray = "#343E44";
      '';
      description = ''
        Configuration written to {file}`$XDG_CONFIG_HOME/nap/config.yaml`.

        See <https://github.com/isabelroses/nap> for the documentation.
      '';
    };
  };

  config =
    let
      cfg = config.programs.nap;
    in
    mkIf cfg.enable {
      home.packages = [ cfg.package ];

      xdg.configFile = {
        "nap/config.yaml" = mkIf ((cfg.colors != {}) || (cfg.settings != {})) {
          source = (settingsFormat.generate "nap-config.yaml" (cfg.colors //
                cfg.settings));
        };
      };
    };
}
