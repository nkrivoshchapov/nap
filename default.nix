{ lib, buildGoModule }:
buildGoModule {
  pname = "nap";
  version = "unstable";

  src = ./.;

  vendorHash = "sha256-Iz5+jZs80wWlbQ6pw0/CQHy2gaFm9pT/LdipfW9Hg4o=";

  ldflags = [
    "-s"
    "-w"
  ];

  meta = {
    description = "Code snippets in your terminal 🛌";
    mainProgram = "nap";
    homepage = "https://github.com/isabelroses/nap";
    license = lib.licenses.mit;
    maintainers = with lib.maintainers; [ isabelroses ];
  };
}
