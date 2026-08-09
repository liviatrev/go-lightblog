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
- **New dependency**: `github.com/deepteams/webp` (pure-Go WebP encoder,
  zero CGO, zero dependencies) added for local WebP generation. Supports
  **lossy** VP8 encoding with quality control, producing much smaller
  thumbnails than lossless-only encoders.
- **WYSIWYG Editor**: change editor from Quill JS to SunEditor.

### Fixed

- Removed the unused `path/filepath` import in `utils/cachebuster.go` that
  broke compilation of the `utils` package.
- Replaced `github.com/HugoSmits86/nativewebp` (lossless-only) with
  `github.com/deepteams/webp` (lossy + lossless) for much smaller WebP
  thumbnails.
- `saveImageFile` now writes directly to the destination file instead of
  using temp-file + rename, fixing thumbnail generation failures on WSL /
  mounted filesystems.
- `ResizeAndCacheThumbnail` now saves the JPG variant first and treats WebP
  generation as best-effort, so a WebP failure never prevents thumbnails
  from being generated.
- WebP encoding now flattens alpha transparency onto a white background
  before lossy encoding. WebP lossy stores any alpha in a separate lossless
  ALPH plane, which ballooned file sizes for PNG uploads (e.g. 1.9MB WebP
  vs 89KB JPG). Flattening makes WebP thumbnails comparable to or smaller
  than JPG.
- WebP encoding now uses `PresetPhoto` for photographic content, producing
  better compression for blog cover images.

## [0.0.0] - Initial release

- Initial commit with cache buster implementation.
