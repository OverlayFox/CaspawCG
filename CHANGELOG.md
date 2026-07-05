# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- #33 - `Add media` button now disappears when entering live mode
- fixed reconnection logic to casparCG-server and fixed frontend to relist elements once the server comes available again
- #39 - elements can not be placed next to each other
- #31 - delay is now respected when playing CG elements out
- fixed a bug with WailsV2 where not setting `max-width` locks the max width of the spawning window
- fixed nil pointer when creating Datasource clients
- clear also now clears mixer commands

### Changed

- changed scaling of elements to give the workarea more space
- cleaned up JavaScript front-end

### Added

- #43 - added update cycle system
- added AI disclosure to the README.md
- #45 - added next button to templates and groups
- #38 - allow for direct string input into `custom fields`
- #35 - media element can now be used inside groups
- #32 - any element can now be given a custom name by the user
- #24 - added clearing and playing out elements to multiple channels at once
- #28 - added build and changelog pipelines
- #20 - added media elements
- #11 - properly added datasources into the front end and integrated google sheets as a datasource
- #14 - added mixer commands to the UI and backend
