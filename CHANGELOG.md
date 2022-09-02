# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

## [1.3.1] - 2022-09-02

### Changed

* Omit optional `SegmentationDescriptor` attribute.

## [1.3.0] - 2022-08-22

### Fixed

* Fixed overflow when building with `GOARCH=386`.

### Changed

* Changed encode command to also read from stdin.

## [1.2.1] - 2021-09-09

### Fixed

* Fixed case where unencrypted signals with `alignment_stuffing` returned `ErrBufferUnderflow`.

## [1.2.0] - 2021-07-07

### Fixed

* Fixed nil pointer in `SpliceInsert.TimeSpecifiedFlag`

### Added

* Added additional methods for computing `_flag` values.

## [1.1.0] - 2021-07-07

### Added

* Added `SpliceInfoSection.Table(prefix, indent)` to format tabular output.

* Added methods for computing `_flag` values. 

* Added latest specification and XML schema to /docs 

## [1.0.0] - 2021-07-01

### Added

* Add code from internal repository
