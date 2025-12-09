# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with documentation in this repository.

## Playwright MCP Integration

### Overview

The project is configured with Playwright MCP for automated browser testing and documentation validation. The configuration is in `.mcp.json` and uses chromium in headless, isolated mode.

### Configuration

Located in `.mcp.json`:

```json
{
  "playwright": {
    "command": "npx",
    "args": ["-y", "@playwright/mcp@latest", "--browser", "chromium", "--headless", "--isolated"],
    "env": {
      "PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD": "0"
    }
  }
}
```

### Key Features

- **Browser**: Chromium (automatically installed via npx)
- **Mode**: Headless and isolated (no persistent profile)
- **Auto-installation**: PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=0 ensures chromium installs on first use

### Initial Setup

If you encounter chromium installation issues, manually install:

```bash
npx playwright install chromium
```

### Common Use Cases

**1. Documentation Site Validation**

```bash
# Navigate to local docs and take screenshot
mcp__playwright__browser_navigate(url: "http://localhost:4000/pulumicost-core/")
mcp__playwright__browser_snapshot()
mcp__playwright__browser_take_screenshot(filename: "docs-homepage.png")
```

**2. GitHub Pages Validation**

```bash
# Check deployed documentation
mcp__playwright__browser_navigate(url: "https://rshade.github.io/pulumicost-core/")
mcp__playwright__browser_snapshot()
```

**3. Link Checking and Navigation Testing**

```bash
# Test documentation navigation
mcp__playwright__browser_navigate(url: "https://rshade.github.io/pulumicost-core/")
mcp__playwright__browser_click(element: "User Guide link", ref: "a[href='/guides/user-guide']")
mcp__playwright__browser_snapshot()
```

**4. Form Testing (Future Plugin Integration)**

```bash
# Test interactive documentation features
mcp__playwright__browser_fill_form(fields: [...])
mcp__playwright__browser_click(element: "Submit button", ref: "button[type='submit']")
```

**5. Network Request Monitoring**

```bash
# Monitor API calls in documentation examples
mcp__playwright__browser_navigate(url: "https://rshade.github.io/pulumicost-core/examples/")
mcp__playwright__browser_network_requests()
```

### Troubleshooting

**Issue: "Chromium distribution 'chrome' is not found"**

- Solution: Run `npx playwright install chromium`
- Root cause: Chromium not installed or wrong browser channel specified

**Issue: Hanging on launch**

- Solution: Ensure `--headless` and `--isolated` flags are set in `.mcp.json`
- Check: `PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=0` is set to allow installation

**Issue: Permission denied in WSL**

- Solution: Add `--no-sandbox --disable-setuid-sandbox` to launchOptions if needed
- Note: Already configured in current setup

### Best Practices

1. **Always use snapshots first**: `browser_snapshot()` is faster than screenshots and provides better context
2. **Screenshots for visual verification**: Use `browser_take_screenshot()` for visual regression testing
3. **Network monitoring**: Use `browser_network_requests()` to validate API calls in documentation examples
4. **Cleanup**: Browser instances are automatically cleaned up due to `--isolated` flag
5. **Documentation validation workflow**:
   - Start local docs: `make docs-serve`
   - Navigate: `browser_navigate(url: "http://localhost:4000/pulumicost-core/")`
   - Validate: `browser_snapshot()` and verify content
   - Test links: Click through navigation and verify no 404s

### Integration with CI/CD

For future automated testing, Playwright can be integrated into GitHub Actions:

```yaml
- name: Install Playwright
  run: npx playwright install chromium
- name: Test Documentation
  run: npx playwright test
```

### Actual Usage Examples

Real-world Playwright MCP usage for testing GitHub Pages:

```bash
# Navigate and verify page loads
mcp__playwright__browser_navigate(url: "https://rshade.github.io/pulumicost-core/")

# Take full page screenshot for visual verification
mcp__playwright__browser_take_screenshot(filename: "site-screenshot.png", fullPage: true)

# Check network requests to verify CSS and assets loaded
mcp__playwright__browser_network_requests()
# Returns: All HTTP requests with status codes (useful for debugging 404s)
```

## GitHub Pages and Jekyll Documentation Setup

### Critical Setup Requirements

**1. Entry Point File:**

- GitHub Pages requires `index.md` or `index.html` as the landing page
- Jekyll does NOT automatically convert `README.md` to `index.html`
- Always create an explicit `index.md` file in the docs directory

**2. Jekyll Plugin Dependencies:**

- Plugins must be installed BEFORE Jekyll can use their template tags
- Common error: `Liquid syntax error: Unknown tag 'seo'` means plugin not loaded
- Solution: Either install the plugin or remove the template tag

**3. Custom CSS Integration:**

- Custom stylesheets must be explicitly linked in `_layouts/default.html`
- Path format: `{{ "/assets/css/style.css?v=" | append: site.github.build_revision | relative_url }}`
- SCSS files in `docs/assets/css/style.scss` are automatically processed by Jekyll

**4. Layout and Content Separation:**

- Avoid duplicate H1 headings between layout header and page content
- Layout typically provides site title/header
- Page content should start with introductory text, not repeat the title

### Common Jekyll Build Errors

**Error: `Unknown tag 'seo'`**

- Cause: `jekyll-seo-tag` plugin not installed or not in `_config.yml` plugins list
- Fix: Either add plugin to Gemfile and \_config.yml, or replace `{% seo %}` with manual tags
- Manual alternative:
  ```html
  <title>{{ page.title | default: site.title }}</title>
  <meta name="description" content="{{ page.description | default: site.description }}" />
  ```

**Error: 404 on GitHub Pages**

- Cause: Missing `index.md` or `index.html` in docs directory
- Fix: Create `index.md` with proper frontmatter:
  ```yaml
  ---
  layout: default
  title: Your Title
  description: Your description
  ---
  ```

**Error: No CSS styling on deployed site**

- Cause: Missing stylesheet link in `_layouts/default.html`
- Fix: Add link tag in `<head>` section:
  ```html
  <link rel="stylesheet" href="{{ "/assets/css/style.css?v=" | append: site.github.build_revision | relative_url }}">
  ```
