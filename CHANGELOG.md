# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [0.5.0] - 2019-01-01

### Added
- **[FEATURE] TSDB InfluxDB 1.x**
  - InfluxDB is now supported. Every tests results will be written in InfluxDB (ResponseTime, Status, TestResult)
  - Grafana dashboard for InfluxDB
- **[FEATURE] DNS Caching**
  - Vigie now cache DNS records to avoid multiples and repetitive DNS queries to Nameservers.
- **[FEATURE] Host section in Vigie Config**
  - Add contexts and info for the running Vigie host.
- **[CI] Build & Release**
  - Added some automation scripts with Goreleaser
### Changed
- **[FIX] Several bug fixes**
### Removed
- Clean up code

## [0.4.0] - 2019-12-13

### Added
- **[FEATURE] Pre-Start for long intervals Probes**
  - Probes T > 59 sec will be exec at the start (between 1-10sec) instead of T after the start.
  - This avoid frustration to wait 'Not Defined' Probes being exec.
- **[FEATURE] Slack**
  - Added Slack alerting
- **[EXP] SubTest**
  - Added `SubTest` variable in `ProbeInfo` struct. The goal is to identify the different results within a TestStep in case of multiple A resolution.
- **[CI] Build & Release**
  - Added some automation scripts

### Changed
- **[FIX] Several bug fixes**
### Removed
- .

## [0.3.0] - 2019-11-30

### Added
- Initial Release
### Changed
- .
### Removed
- .
