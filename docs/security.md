# Security Model

VPS Inspector runs on infrastructure that may be exposed to the internet, so the default posture is conservative.

## Defaults

- Listen on `127.0.0.1:8719`.
- Require a bearer token when `VPS_INSPECTOR_AUTH_TOKEN` is set.
- Avoid shell string execution.
- Use context deadlines for all checks.
- Keep check inputs server-defined in the first version.

## Command Execution

Any future external command must pass through `internal/runner`. Commands should be allowlisted by executable name and arguments should be assembled as slices, not shell strings.

## Public Exposure

If binding to `0.0.0.0`, put the service behind HTTPS and set `VPS_INSPECTOR_AUTH_TOKEN` to a long random value.

