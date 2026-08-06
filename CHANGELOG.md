# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **WebP thumbnail output**: `ResizeAndCacheThumbnail` now generates both JPEG
  and WebP variants of every resized thumbnail, so templates can serve modern
  WebP images with a JPEG fallback for older browsers.
- **`/api/thumb` format parameter**: the thumbnail proxy endpoint accepts a
  `f` query parameter (`f=jpg` or `f=webp`) to request a specific output
  format.
- **`thumb` template helper**: now returns both WebP and JPG URLs instead of a
  single URL, enabling `<picture>` elements with WebP-first loading in
  `home.html`, `post.html`, `archive.html`, and `search.html`.

### Changed

- **Dynamic upload/thumbnail storage path**: `ProcessUpload` and
  `ResizeAndCacheThumbnail` no longer hardcode `./public`. Storage now
  resolves against the `-a` / `--public` flag value via the new
  `config.PublicPath` global.
- **ImageKit CDN URLs**: the `thumb` helper now emits explicit `f-jpg` and
  `f-webp` ImageKit transformations for the respective fallback variants.
- **New dependency**: `github.com/HugoSmits86/nativewebp` (pure-Go WebP
  encoder, no cgo/libwebp required) added for local WebP generation.

### Fixed

- Removed the unused `path/filepath` import in `utils/cachebuster.go` that
  broke compilation of the `utils` package.

## [0.0.0] - Initial release

- Initial commit with cache buster implementation.
