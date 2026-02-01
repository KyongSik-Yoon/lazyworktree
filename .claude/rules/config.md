---
paths:
  - "internal/config/**/*.go"
---

# Configuration Rules

When working with configuration code:

- **Viper integration**: Use Viper for configuration management.
- **Default values**: Provide sensible defaults for all configuration options.
- **Environment variables**: Support `LAZYWORKTREE_*` environment variable overrides.
- **Config file locations**: Follow XDG Base Directory specification.
- **CLI flag mapping**: Ensure config options map to corresponding CLI flags.
- **Validation**: Validate configuration values at load time.
- **Documentation**: Update README when adding new configuration options.
