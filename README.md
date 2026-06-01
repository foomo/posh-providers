[![Go Report Card](https://goreportcard.com/badge/github.com/foomo/posh-providers?style=flat-square)](https://goreportcard.com/report/github.com/foomo/posh-providers)
[![GoDoc](https://img.shields.io/badge/GoDoc-✓-informational.svg?style=flat-square&logo=go)](https://godoc.org/github.com/foomo/posh-providers)
[![Coverage](https://img.shields.io/codecov/c/github/foomo/posh-providers?style=flat-square&logo=github)](https://app.codecov.io/gh/foomo/posh-providers)
[![GitHub Stars](https://img.shields.io/github/stars/foomo/posh-providers.svg?style=flat-square&logo=github)](https://github.com/foomo/posh-providers)

<p align="center">
  <img alt="posh-providers" src="docs/public/logo.png" width="400" height="400"/>
</p>

# Project Oriented SHELL (posh) providers

This repository is a collection of provider modules for [posh](https://github.com/foomo/posh), a project-scoped
interactive shell. Each provider plugs an external tool or service — such as `kubectl`, `gcloud`, `squadron`, or
`1Password` — into the posh prompt, exposing its commands, suggestions, and health checks alongside the rest of your
project tooling.

The repo is organised as a Go multi-module workspace: every provider lives in its own module under `<vendor>/<tool>/`
and can be pulled in independently by downstream posh projects. See each provider's own `README.md` for wiring snippets
and sample configuration.

## How to Contribute

Contributions are welcome! Please read the [contributing guide](docs/CONTRIBUTING.md).

![Contributors](https://contributors-table.vercel.app/image?repo=foomo/posh-providers&width=50&columns=15)

## License

Distributed under MIT License, please see the [license](LICENSE) file within the code for more details.

_Made with ♥ [foomo](https://www.foomo.org) by [bestbytes](https://www.bestbytes.com)_
