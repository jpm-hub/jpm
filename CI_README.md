# CI/CD with GitHub Actions

This project uses GitHub Actions for automated building, testing, and releasing across all supported platforms.

## Workflows

### 1. Build and Test (`build.yml`)
**Triggers:** Pull requests, pushes to non-main branches

**What it does:**
- Runs tests on Ubuntu, macOS, and Windows
- Builds binaries for all platforms (Linux, macOS, Windows) on both AMD64 and ARM64
- Ensures code quality and cross-platform compatibility

**Use cases:**
- Development workflow
- Pull request validation
- Feature branch testing

### 2. Build and Release (`release.yml`)
**Triggers:** Push to main branch, manual dispatch

**What it does:**
- Builds binaries for all platforms
- Creates GitHub releases automatically
- Uploads platform-specific archives
- Extracts version from `common/templates.go`

**Use cases:**
- Production releases
- Automated versioning
- Distribution to users

## Supported Platforms

| Platform | Architecture | Binary Name | Archive Format |
|----------|--------------|-------------|----------------|
| Linux    | AMD64        | `jpm`       | `.tar.gz`      |
| Linux    | ARM64        | `jpm`       | `.tar.gz`      |
| macOS    | Intel        | `jpm`       | `.tar.gz`      |
| macOS    | Apple Silicon| `jpm`       | `.tar.gz`      |
| Windows  | AMD64        | `jpm.exe`   | `.zip`         |
| Windows  | ARM64        | `jpm.exe`   | `.zip`         |

## How It Works

### Version Management
The release workflow automatically extracts the version from `common/templates.go`:
```go
// In common/templates.go
println(`   ____________  ___
  |_  | ___ \  \/  |  version: 0.0.1  // ← This version is used
    | | |_/ / .  . |  The simpler
    | |  __/| |\/| |  package manager
/\__/ / |   | |  | |  for your
\____/\_|   \_|  |_/  JVM project`)
```

### Build Process
1. **Matrix Strategy**: Builds for all 6 platform combinations simultaneously
2. **Cross-compilation**: Uses Go's built-in cross-compilation with `GOOS` and `GOARCH`
3. **Optimization**: Applies `-ldflags="-s -w"` for smaller binaries
4. **CGO Disabled**: Ensures static linking and maximum compatibility

### Release Process
1. **Artifact Collection**: Downloads all built binaries
2. **Version Detection**: Automatically finds current version
3. **Release Creation**: Creates GitHub release with proper tagging
4. **Asset Upload**: Attaches all platform-specific archives

## Manual Release

To manually trigger a release:
1. Go to Actions tab in GitHub
2. Select "Build and Release" workflow
3. Click "Run workflow"
4. Select branch (usually main)
5. Click "Run workflow"

## Customization

### Adding New Platforms
To add support for a new platform, update the matrix in both workflows:

```yaml
- os: newplatform
  arch: newarch
  goos: newgoos
  goarch: newgoarch
  binary_name: jpm
  archive_name: jpm-newplatform-newarch.tar.gz
```

### Changing Go Version
Update the `GO_VERSION` environment variable in both workflows:

```yaml
env:
  GO_VERSION: '1.22'  # Change this
```

### Modifying Build Flags
Update the build command in the build steps:

```yaml
go build -ldflags="-s -w -X main.Version=${{ steps.version.outputs.version }}" -o ${{ matrix.binary_name }} jpm.go
```

## Troubleshooting

### Build Failures
- Check Go version compatibility
- Verify all dependencies are available
- Check for platform-specific code issues

### Release Failures
- Ensure `GITHUB_TOKEN` has proper permissions
- Check version format in `templates.go`
- Verify all artifacts were built successfully

### Performance Issues
- Matrix builds run in parallel
- Consider using `fail-fast: false` for debugging
- Use `runs-on: ubuntu-latest` for fastest builds

## Security

- **CGO Disabled**: Prevents dynamic linking vulnerabilities
- **Static Binaries**: Self-contained executables
- **Minimal Dependencies**: Only essential Go runtime included
- **Signed Releases**: GitHub automatically signs releases

## Monitoring

- **Build Status**: Check Actions tab for real-time status
- **Release History**: View all releases in Releases tab
- **Artifact Storage**: Build artifacts are stored for 90 days
- **Logs**: Detailed logs available for debugging
