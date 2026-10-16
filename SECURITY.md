# Security Policy

Bourbon is designed from the ground up with strict security invariants, zero third-party runtime dependencies in core, and standard library cryptographic primitives.

---

## Reporting a Vulnerability

If you discover a security vulnerability in Bourbon, please report it via private disclosure rather than opening a public issue.

- **Email**: `security@bourbon.dev`
- **Response Target**: Within 48 hours
- **Advisory Protocol**: Coordinated disclosure with CVE assignment upon fix release.

---

## Security Invariants in Bourbon Core

1. **Zero Runtime Dependencies**: The core module imports only the Go standard library, drastically minimizing the attack surface and supply-chain vulnerabilities.
2. **Timing-Safe Authentication**: Token signatures, CSRF secrets, and password hashes are compared using constant-time comparison primitives (`crypto/subtle.ConstantTimeCompare`, `hmac.Equal`).
3. **Strict Path Traversal Defenses**: All file storage operations (`storage/localfs`, `middleware/static`) reject directory traversal sequences (`..`, null bytes, symlink escapes).
4. **RFC 7807 Safe Errors**: Unhandled panic stack traces and internal database errors are masked behind structured Problem Details responses in production.
5. **Memory & Connection Protection**: Built-in HTTP timeouts, rate limiting tokens, and body size limits prevent Slowloris and resource exhaustion attacks.
