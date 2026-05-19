# Third-Party Tools

ratnosint7 orchestrates 13 open-source subdomain enumeration tools. Each runs independently; ratnosint7 deduplicates & normalizes output.

## OSINT / Passive Only (Safe, no DNS to target)

| Tool | Type | URL | License |
|------|------|-----|---------|
| **subfinder** | Go | https://github.com/projectdiscovery/subfinder | [License](https://github.com/projectdiscovery/subfinder/blob/main/LICENSE.md) |
| **assetfinder** | Go | https://github.com/tomnomnom/assetfinder | [MIT](https://github.com/tomnomnom/assetfinder/blob/master/LICENSE) |
| **sublist3r** | Python | https://github.com/aboul3la/Sublist3r | [GPL-3.0](https://github.com/aboul3la/Sublist3r/blob/master/LICENSE) |
| **subscraper** | Python | https://github.com/m8sec/subscraper | [MIT](https://github.com/m8sec/subscraper/blob/main/LICENSE) |
| **turbolist3r** | Python | https://github.com/fleetcaptain/Turbolist3r | [GPL-2.0](https://github.com/fleetcaptain/Turbolist3r/blob/master/LICENSE) |
| **dome** | Python | https://github.com/v4d1/Dome | [GPL-3.0](https://github.com/v4d1/Dome/blob/main/LICENSE) |
| **as3nt** | Python | https://github.com/cinerieus/as3nt | [BSD-3-Clause](https://github.com/cinerieus/as3nt/blob/main/LICENSE) |
| **substr3am** | Python | https://github.com/nexxai/Substr3am | [MIT](https://github.com/nexxai/Substr3am/blob/main/LICENSE) |

## Passive + Active (OSINT + DNS Brute-Force)

| Tool | Type | URL | License |
|------|------|-----|---------|
| **amass** | Go | https://github.com/owasp-amass/amass | [Apache 2.0](https://github.com/owasp-amass/amass/blob/master/LICENSE) |
| **findomain** | Rust | https://github.com/Findomain/Findomain | [GPL-3.0](https://github.com/Findomain/Findomain/blob/master/LICENSE) |
| **sudomy** | Python | https://github.com/Screetsec/Sudomy | [MIT](https://github.com/Screetsec/Sudomy/blob/master/LICENSE) |
| **tugarecon** | Python 3.12 | https://github.com/skynet0x01/tugarecon | [GPL-3.0](https://github.com/skynet0x01/tugarecon/blob/master/LICENSE) |

## Active Only / DNS Validation (Requires Authorization)

| Tool | Type | URL | License |
|------|------|-----|---------|
| **dnsx** | Go | https://github.com/projectdiscovery/dnsx | [License](https://github.com/projectdiscovery/dnsx/blob/main/LICENSE.md) |

## Installation

Tools are external binaries/scripts installed via `ratnosint7 update-tools`:

```bash
./ratnosint7 update-tools
# [1/13] installing subfinder …
#          ✓  subfinder installed successfully
# Summary: Success: 13 / 13
```

Some require pre-installation of runtimes:
- **Python tools**: Require `python3`, `pip3` (most tools)
- **tugarecon**: Requires **`python3.12`** on PATH (`brew install python@3.12`); dependencies go into `~/.ratnosint7/tools/tugarecon/.venv` (not system-wide pip)
- **Rust tools** (findomain): Requires `cargo` (https://rustup.rs/)
- **Go tools**: Require Go 1.21+

## Attribution & Licensing

Each tool has its own license (check their GitHub repos). ratnosint7 is Apache 2.0 and orchestrates them via subprocess execution—not derivative works.

Always check each tool's license before commercial use.
