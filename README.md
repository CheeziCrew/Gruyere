# Gruyere

![cheezi_gruyere](https://github.com/user-attachments/assets/5645a782-cf6f-493e-9395-39769c0e5f95)


Age your APIs gracefully. Interactive TUI for generating API changelogs.

## Status

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=CheeziCrew_Gruyere&metric=alert_status)](https://sonarcloud.io/summary/overall?id=CheeziCrew_Gruyere)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=CheeziCrew_Gruyere&metric=reliability_rating)](https://sonarcloud.io/summary/overall?id=CheeziCrew_Gruyere)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=CheeziCrew_Gruyere&metric=security_rating)](https://sonarcloud.io/summary/overall?id=CheeziCrew_Gruyere)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=CheeziCrew_Gruyere&metric=sqale_rating)](https://sonarcloud.io/summary/overall?id=CheeziCrew_Gruyere)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=CheeziCrew_Gruyere&metric=vulnerabilities)](https://sonarcloud.io/summary/overall?id=CheeziCrew_Gruyere)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=CheeziCrew_Gruyere&metric=bugs)](https://sonarcloud.io/summary/overall?id=CheeziCrew_Gruyere)

## Install

```sh
go install github.com/CheeziCrew/gruyere@latest
```

Or grab a binary from [Releases](https://github.com/CheeziCrew/gruyere/releases).

## Usage

### TUI (default)

```sh
gruyere
```

### CLI

```sh
gruyere <base-branch> <feature-branch>
gruyere -o changelog.md main feature/UF-123
gruyere --prepend main feature/UF-123
```

## What it does

1. Finds `openapi.yaml` in your repo
2. Fetches the spec from both branches via `git show`
3. Compares endpoints (added/removed) and schemas (added/removed/modified fields)
4. Generates a markdown changelog

## Part of the Swiss Tools family

Built on [curd](https://github.com/CheeziCrew/Curd) — the shared TUI component library.
