# nap

nap is a highly opinionated fork of nap

To move to nap just rename all your previously defined `NAP_` env vars to `NAP_`

<img width="1200" alt="nap" src="https://user-images.githubusercontent.com/42545625/202545409-eb53f92a-233a-4f78-b598-a59c65248ad3.png">

<sub><sub>z</sub></sub><sub>z</sub>z

Zzz is a code snippet manager for your terminal. Create and access new snippets
quickly with the command-line interface or browse, manage, and organize them with the
text-user interface. Keep your code snippets safe, sound, and well-rested in your terminal.

<br />

<p align="center">
<img width="1000" src="https://user-images.githubusercontent.com/42545625/202577549-f2e0887a-b740-41f4-9408-c2f53673503f.gif" />
</p>

<br />

## Text-based User Interface

Launch the interactive interface:

```bash
nap
```

<img width="1000" src="./tapes/nap-demo.gif" />

<details>

<summary>Key Bindings</summary>

<br />

| Action                               | Key                            |
| :----------------------------------- | :----------------------------- |
| Create a new snippet                 | <kbd>n</kbd>                   |
| Edit selected snippet (in `$EDITOR`) | <kbd>e</kbd>                   |
| Copy selected snippet to clipboard   | <kbd>c</kbd>                   |
| Paste clipboard to selected snippet  | <kbd>p</kbd>                   |
| Delete selected snippet              | <kbd>x</kbd>                   |
| Move selected snippet up             | <kbd>K</kbd>                   |
| Move selected snippet down           | <kbd>J</kbd>                   |
| Rename selected snippet              | <kbd>r</kbd>                   |
| Rename selected folder               | <kbd>R</kbd>                   |
| Move to next pane                    | <kbd>l</kbd> <kbd>→</kbd>      |
| Move to previous pane                | <kbd>h</kbd> <kbd>←</kbd>      |
| Search for snippets                  | <kbd>/</kbd>                   |
| Toggle help                          | <kbd>?</kbd>                   |
| Quit application                     | <kbd>q</kbd> <kbd>ctrl+c</kbd> |

</details>

## Command Line Interface

Create new snippets:

```bash
# Quick save an untitled snippet.
nap < main.go

# From a file, specify Notes/ folder and Go language.
nap Notes/FizzBuzz.go < main.go

# Save some code from the internet for later.
curl https://example.com/main.go | nap Notes/FizzBuzz.go

# Works great with GitHub gists
gh gist view 4ff8a6472247e6dd2315fd4038926522 | nap
```

<img width="600" src="./tapes/nap-save.gif" />

Output saved snippets:

```bash
# Fuzzy find snippet.
nap fuzzy

# Write snippet to a file.
nap go/boilerplate > main.go

# Copy snippet to clipboard.
nap foobar | pbcopy
nap foobar | xclip
nap foobar | wl-copy
```

<img width="600" src="./tapes/fuzzy-find.gif" />

List snippets:

```bash
nap list
```

<img width="600" src="./tapes/nap-list.gif" />

Fuzzy find a snippet (with [Gum](https://github.com/charmbracelet/gum)).

```bash
nap $(nap list | gum filter)
```

## Installation

###### Install with nix (using flake + home manager):

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    nap.url = "github:isabelroses/nap";
  };

  outputs = { self, nixpkgs, home-manager, nap }: {
    homeConfigurations."user@hostname" = home-manager.lib.homeManagerConfiguration {
      modules = [
        home-manager.homeManagerModules.default
        nap.homeManagerModules.default
        {
          programs.nap = {
            settings = {
              home = "~/.nap";
            };
            colors = {
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
            };
          };
        }
      ];
    };
  }
}
```

###### Install with Go:

```sh
go install github.com/isabelroses/nap@main
```

Or download a binary from the [releases](https://github.com/isabelroses/nap/releases).

## Customization

Zzz is customized through a configuration file located at `NAP_CONFIG` (`$XDG_CONFIG_HOME/nap/config.yaml`).

```yaml
# Configuration
home: ~/.nap
default_language: go
theme: nord

# Colors
background: "0"
foreground: "7"
primary_color: "#AFBEE1"
primary_color_subdued: "#64708D"
green: "#527251"
bright_green: "#BCE1AF"
bright_red: "#E49393"
red: "#A46060"
gray: "240"
textinvert: "#373B41"
```

The configuration file can be overridden through environment variables:

```bash
# Configuration
export NAP_CONFIG="~/.nap/config.yaml"
export NAP_HOME="~/.nap"
export NAP_DEFAULT_LANGUAGE="go"
export NAP_THEME="nord"

# Colors
export NAP_PRIMARY_COLOR="#AFBEE1"
export NAP_RED="#A46060"
export NAP_GREEN="#527251"
export NAP_FOREGROUND="7"
export NAP_BACKGROUND="0"
export NAP_GRAY="240"
export NAP_TEXTINVERT="#373B41"
```

<br />

<p align="center">
  <img
    width="1000"
    alt="image"
    src="https://user-images.githubusercontent.com/42545625/202867429-5bcf8fae-5dd7-478c-b958-638aa5765d97.png"
  />
</p>

## License

[MIT](https://github.com/isabelroses/nap/blob/master/LICENSE)

## Feedback

I'd love to hear your feedback on improving `nap`.

Feel free to reach out via:

- [Contact](https://isabel.contact)
- [Email](mailto:isabel@isabelroses.com)
- [GitHub issues](https://github.com/isabelroses/nap/issues/new)

---

<sub><sub>z</sub></sub><sub>z</sub>z
