# Security Policy

## Reporting a Vulnerability

**DO NOT** open a public GitHub issue for security vulnerabilities. Public disclosure can allow attackers to exploit the vulnerability before a fix is available.

Report vulnerabilities privately via [GitHub Security Advisories](https://github.com/Ratnadeepdeyroy/ratnosint7/security/advisories/new) on this repository. Do not open a public issue.

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if you have one)

**Response timeline:**
- Initial acknowledgment: Within 24 hours
- Status update: Within 7 days
- Patch release: Within 30 days (or explanation if longer)

Your report will be kept confidential until a fix is released.

## Scope

### In Scope
- Authentication/authorization bypass in orchestration engine
- Insecure default configurations
- Information disclosure in logs/output
- Goroutine leaks or resource exhaustion
- Context cancellation failures
- Unsafe channel operations
- Command injection in tool execution
- Path traversal in output/cache directories
- Race conditions in concurrent access patterns

### Out of Scope
- Vulnerabilities in third-party tools (report directly to tool authors)
- Denial of service via intentional rate limiting bypass (feature, not bug)
- Network-based attacks (scan traffic patterns, IDS evasion)
- Social engineering or phishing
- Physical security
- Issues requiring specific network/ISP cooperation

## Security Best Practices

When using ratnosint7:

1. **Never scan without authorization** — Always have written permission
2. **Review tool behavior** — Understand what each tool sends on the network
3. **Test on owned domains first** — Verify output quality + legal compliance
4. **Check local policies** — Verify network/ISP allows enumeration traffic
5. **Use passive mode by default** — Active (DNS brute-force) has higher legal risk
6. **Monitor logs** — Check `~/.ratnosint7/logs/` for tool errors
7. **Keep tools updated** — Run `./ratnosint7 update-tools` regularly
8. **Review third-party licenses** — Ensure compliance with tool licenses

## Known Limitations

- **No secrets scanning**: ratnosint7 does not scan for exposed credentials in results
- **No rate-limiting evasion**: Active scans send consistent DNS traffic patterns
- **No proxy support**: Direct network access only
- **No authentication for APIs**: Tools use public OSINT sources

## Deprecation & End-of-Life

- **Go versions**: Only Go 1.21+ supported. Older versions receive no patches.
- **Third-party tools**: If a tool is abandoned, ratnosint7 will remove it (announce 1 release in advance)
- **Deprecated features**: 12-month deprecation window before removal

## Security Updates

Security patches are released as soon as possible after confirmation:
- **Critical**: ASAP release
- **High**: Within 2 weeks
- **Medium**: Within 1 month
- **Low**: Next regular release

All security releases include a public advisory on GitHub Releases.
