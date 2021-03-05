# tricktool

tricktool is a companion utility app for Trickster, and is intended to hold a collection of miscellaneous tools for managing your Trickster deployments.

## Usage

Currently tricktool offers one command, `upgrade-config`, which will upgrade a Trickster 1.x TOML configuration to a Trickster 2.0 YAML configuration. The command accepts 1 argument: the path to the source file; and will output the upgraded config to stdout, so long as there are no errors in the TOML file. Usage is as follows:

```bash
tricktool upgrade-config /etc/trickster.conf > /etc/trickster.yaml
```

## Downloading

Download tricktool from the [Releases](https://github.com/tricksterproxy/tricktool/releases) page, or (if you have golang installed on your device,) run `go get github.com/tricksterproxy/tricktool`, and it will be built and placed in your `$GOBIN`.

## Special Note About the Project

This project is brand new, so while we hope it will work flawlessly for everyone, and has worked for everything we've thrown at it, we're very eager to refine it further while Trickster 2.0 is in Beta. If you are having any problems with `upgrade-config`, please file an [issue](https://github.com/tricksterproxy/tricktool/issues), and if possible, attach a sanitized version of your 1.x configuration.

Likewise, if there are any other useful tools you'd like to see that would facilitate your Trickster deployment, file an [issue](https://github.com/tricksterproxy/tricktool/issues) here with a feature request.

[Contributions](https://github.com/tricksterproxy/tricktool/CONTRIBUTING.md) are always welcome and most appreciated. We work very hard to review PR's within 24 hours and credit the work to the contributor in the next release's notes.
