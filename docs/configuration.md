# Configuration Guide

Medha uses a JSON configuration file located at `~/.medha/configs/config.json`. If no configuration file exists, Medha uses sensible defaults.

## Configuration File Location

```
~/.medha/
├── configs/
│   └── config.json    # Main configuration file
├── db/
│   └── medha.db       # SQLite database (if using SQLite)
└── store/
    └── medha-{user}/  # Git repositories
```

## Complete Configuration Reference

```json
{
  "server": {
    "host": "localhost",
    "port": 8080,
    "tls": {
      "enabled": false,
      "cert_file": "/path/to/cert.pem",
      "key_file": "/path/to/key.pem"
    }
  },
  "database": {
    "type": "sqlite",
    "sqlite_path": "~/.medha/db/medha.db",
    "postgres_dsn": "postgres://user:pass@localhost:5432/medha?sslmode=disable"
  },
  "auth": {
    "type": "local"
  },
  "saml": {
    "entity_id": "https://your-domain.com",
    "acs_url": "https://your-domain.com/saml/acs",
    "metadata_url": "https://your-domain.com/saml/metadata",
    "idp_metadata": "https://idp.example.com/metadata",
    "certificate": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
    "private_key": "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
    "provider": "okta"
  },
  "git": {
    "default_branch": "main",
    "sync_interval_minutes": 60
  },
  "security": {
    "encryption_key": "",
    "token_ttl_hours": 24
  }
}
```

## Configuration Sections

### Server Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `server.host` | string | `"localhost"` | Host to bind the HTTP server |
| `server.port` | int | `8080` | Port to listen on (1-65535) |
| `server.tls.enabled` | bool | `false` | Enable HTTPS |
| `server.tls.cert_file` | string | `""` | Path to TLS certificate file |
| `server.tls.key_file` | string | `""` | Path to TLS private key file |

**Note:** Server configuration only applies when running in HTTP mode (`--http` flag).

### Database Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `database.type` | string | `"sqlite"` | Database type: `"sqlite"` or `"postgres"` |
| `database.sqlite_path` | string | `"~/.medha/db/medha.db"` | Path to SQLite database file |
| `database.postgres_dsn` | string | `""` | PostgreSQL connection string |

**SQLite (Development/Local)**
```json
{
  "database": {
    "type": "sqlite",
    "sqlite_path": "~/.medha/db/medha.db"
  }
}
```

**PostgreSQL (Production)**
```json
{
  "database": {
    "type": "postgres",
    "postgres_dsn": "postgres://user:password@localhost:5432/medha?sslmode=require"
  }
}
```

### Authentication Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `auth.type` | string | `"local"` | Authentication type: `"local"` or `"saml"` |

**Local Authentication (Development)**

Uses system username from `whoami` or `ACCESSING_USER` environment variable.

```json
{
  "auth": {
    "type": "local"
  }
}
```

**SAML Authentication (Production)**

Enables enterprise SSO with providers like Okta or DUO.

```json
{
  "auth": {
    "type": "saml"
  }
}
```

### SAML Configuration

Required when `auth.type` is `"saml"`.

| Field | Type | Description |
|-------|------|-------------|
| `saml.entity_id` | string | Your service provider entity ID (typically your domain) |
| `saml.acs_url` | string | Assertion Consumer Service URL for SAML responses |
| `saml.metadata_url` | string | URL where your SP metadata is served |
| `saml.idp_metadata` | string | URL to your IdP's metadata XML |
| `saml.certificate` | string | PEM-encoded X.509 certificate for signing |
| `saml.private_key` | string | PEM-encoded private key for signing |
| `saml.provider` | string | IdP provider: `"okta"` or `"duo"` |

**Okta Setup Example**
```json
{
  "auth": {
    "type": "saml"
  },
  "saml": {
    "entity_id": "https://medha.yourcompany.com",
    "acs_url": "https://medha.yourcompany.com/saml/acs",
    "metadata_url": "https://medha.yourcompany.com/saml/metadata",
    "idp_metadata": "https://yourcompany.okta.com/app/exk123/sso/saml/metadata",
    "certificate": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
    "private_key": "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
    "provider": "okta"
  }
}
```

### Git Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `git.default_branch` | string | `"main"` | Default branch for new repositories |
| `git.sync_interval_minutes` | int | `60` | Auto-sync interval in minutes (minimum: 1) |

### Security Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `security.encryption_key` | string | `""` | 32-character key for encrypting PAT tokens |
| `security.token_ttl_hours` | int | `24` | Authentication token lifetime in hours |

**Important:** The `encryption_key` is typically provided via the `ENCRYPTION_KEY` environment variable rather than in the config file to avoid storing secrets in plain text.

## Environment Variables

Environment variables take precedence over config file values:

| Variable | Description |
|----------|-------------|
| `ENCRYPTION_KEY` | 32-character encryption key for PAT tokens (required) |
| `ACCESSING_USER` | Username override for local auth (with `--with-accessinguser` flag) |
| `MEDHA_HOME` | Override default data directory (default: `~/.medha`) |

## Minimal Configurations

### Development (Defaults)

No configuration file needed. Medha uses sensible defaults:
- SQLite database at `~/.medha/db/medha.db`
- Local authentication
- 60-minute sync interval

### Production with PostgreSQL

```json
{
  "database": {
    "type": "postgres",
    "postgres_dsn": "postgres://medha:secret@db.example.com:5432/medha?sslmode=require"
  },
  "auth": {
    "type": "saml"
  },
  "saml": {
    "entity_id": "https://medha.example.com",
    "acs_url": "https://medha.example.com/saml/acs",
    "metadata_url": "https://medha.example.com/saml/metadata",
    "idp_metadata": "https://idp.example.com/metadata",
    "certificate": "...",
    "private_key": "...",
    "provider": "okta"
  },
  "server": {
    "host": "0.0.0.0",
    "port": 443,
    "tls": {
      "enabled": true,
      "cert_file": "/etc/medha/tls/cert.pem",
      "key_file": "/etc/medha/tls/key.pem"
    }
  }
}
```

### Docker Container

When using Docker, configuration is typically provided via environment variables and volume mounts:

```json
{
  "mcpServers": {
    "medha": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-v", "/home/user/.medha:/home/medha/.medha",
        "-e", "ENCRYPTION_KEY=your-32-char-key-here",
        "-e", "ACCESSING_USER=username",
        "tejzpr/medha-mcp",
        "--with-accessinguser"
      ]
    }
  }
}
```

## Validation Rules

Medha validates configuration on startup:

- `auth.type` must be `"local"` or `"saml"`
- `database.type` must be `"sqlite"` or `"postgres"`
- `database.sqlite_path` required when type is `"sqlite"`
- `database.postgres_dsn` required when type is `"postgres"`
- `server.port` must be between 1 and 65535
- `git.sync_interval_minutes` must be at least 1
- `security.token_ttl_hours` must be at least 1
- When `auth.type` is `"saml"`: `entity_id`, `acs_url`, and `idp_metadata` are required

## Generating Encryption Keys

Generate a secure 32-character encryption key:

```bash
# Using openssl
openssl rand -base64 32

# Using /dev/urandom
head -c 32 /dev/urandom | base64
```

## Troubleshooting

**Config file not found**
- Medha will use defaults if `~/.medha/configs/config.json` doesn't exist
- Run `medha --http` once to auto-create the directory structure

**Database connection failed**
- For SQLite: Ensure the directory exists and is writable
- For PostgreSQL: Verify the DSN and network connectivity

**SAML authentication issues**
- Verify IdP metadata URL is accessible
- Check certificate/key pair match
- Ensure ACS URL is correctly configured in your IdP
