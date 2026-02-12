# Changelog

All notable changes to the Reforge Go SDK will be documented in this file.

## [1.2.1] - 2025-02-12

### Fixed

- **ReturnError mode re-blocking on every SDK call after init timeout** ([#18](https://github.com/ReforgeHQ/sdk-go/pull/18)) — When `OnInitializationFailure = ReturnError` (the default), the `initializationComplete` channel was never closed after a timeout. This caused every subsequent SDK call to re-block for the full `InitializationTimeoutSeconds` (default 10s) indefinitely. The channel is now closed on the first timeout regardless of failure mode.

## [1.2.0] - 2025-12-20

### Fixed

- **SDK key environment variable not used for SSE and telemetry** ([#16](https://github.com/ReforgeHQ/sdk-go/pull/16)) — When using `REFORGE_BACKEND_SDK_KEY` env var instead of `WithSdkKey()`, SSE live updates and telemetry submission used empty authentication. The SDK key is now resolved once at startup before any components are initialized.

## [1.1.0] - 2025-10-31

### Added

- **Dynamic log level management** with real-time SSE updates ([#10](https://github.com/ReforgeHQ/sdk-go/pull/10))
- **slog integration** — `ReforgeHandler` and `ReforgeLeveler` for stdlib `log/slog`
- **zerolog integration** — separate module at `integrations/zerolog` ([#10](https://github.com/ReforgeHQ/sdk-go/pull/10))
- **zap integration** — separate module at `integrations/zap` ([#10](https://github.com/ReforgeHQ/sdk-go/pull/10))
- **charmbracelet/log integration** — separate module at `integrations/charmbracelet` ([#13](https://github.com/ReforgeHQ/sdk-go/pull/13))

### Changed

- Logger integrations restructured as **separate Go modules** under `integrations/` for zero-bloat imports ([#14](https://github.com/ReforgeHQ/sdk-go/pull/14))

### Fixed

- Proto compilation module path ([#9](https://github.com/ReforgeHQ/sdk-go/pull/9))
- Proto registry conflict with prefab-cloud-go ([#11](https://github.com/ReforgeHQ/sdk-go/pull/11), [#12](https://github.com/ReforgeHQ/sdk-go/pull/12))

## [1.0.0] - 2025-10-10

### Added

- Pluggable `EnvLookup` interface for embedded scenarios ([#8](https://github.com/ReforgeHQ/sdk-go/pull/8))
- Optional `EnvId` field on `ConfigMatch` ([#6](https://github.com/ReforgeHQ/sdk-go/pull/6))

### Changed

- SDK key environment variable changed from `REFORGE_SDK_KEY` to `REFORGE_BACKEND_SDK_KEY` ([#7](https://github.com/ReforgeHQ/sdk-go/pull/7))
- Minimum Go version set to 1.23 ([#5](https://github.com/ReforgeHQ/sdk-go/pull/5))
- Complete package rename from prefab-cloud-go to ReforgeHQ/sdk-go ([#2](https://github.com/ReforgeHQ/sdk-go/pull/2))

### Fixed

- SSE phantom empty events bug ([#3](https://github.com/ReforgeHQ/sdk-go/pull/3))
- Empty response protection for HTTP config endpoint ([#4](https://github.com/ReforgeHQ/sdk-go/pull/4))

### Security

- Bump golang.org/x/net from 0.36.0 to 0.38.0 ([#1](https://github.com/ReforgeHQ/sdk-go/pull/1))
