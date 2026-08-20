{
  buildGoModule,
  lib,
  ffmpeg-headless,
}:

let
  version = builtins.readFile ./VERSION;
in
buildGoModule rec {
  pname = "srtrelay";
  inherit version;
  env.CGO_ENABLED = 0;
  src = ./.;
  vendorHash = "sha256-dd6CVSSP8C1J2L8DjYFM4GRJmNV8PxazyN4fCiwwzdo=";
  nativeCheckInputs = [ ffmpeg-headless ];
  meta = {
    description = "SRT relay server for distributing media streams to multiple clients";
    homepage = "https://github.com/voc/srtrelay";
    platforms = lib.platforms.linux;
    license = lib.licenses.mit;
    maintainers = with lib.maintainers; [
      ischluff
    ];
    mainProgram = "srtrelay";
  };
}
