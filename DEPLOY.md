# Deployment Procedure

## Release Checklist

Before releasing a new version:

1. **Test the build locally**
   ```bash
   go build -o claude-code-switcher.exe -ldflags="-H windowsgui" .
   ```

2. **Run tests**
   ```bash
   go test ./...
   ```

3. **Test the executable manually** - verify the fix/feature works

## Release Steps

### 1. Update CHANGELOG.md

Add a new section at the top (below the header):

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- New features

### Fixed
- Bug fixes

### Changed
- Changes to existing functionality
```

### 2. Update README.md

Update the version number in the badge line:

```markdown
> **Latest Version**: X.Y.Z | [See What's New](CHANGELOG.md)
```

### 3. Commit and Tag

```bash
# Stage and commit your changes (if not already committed)
git add .
git commit -m "Description of changes"

# Commit the version bump
git add CHANGELOG.md README.md
git commit -m "Release X.Y.Z"

# Create the tag (NO 'v' prefix - required for GitHub Actions)
git tag -a X.Y.Z -m "Release X.Y.Z"

# Push commits and tag
git push
git push origin X.Y.Z
```

## Tag Format

The GitHub Actions release workflow triggers on tags matching `[0-9]+.[0-9]+.[0-9]+`.

- Correct: `0.1.1`, `1.0.0`, `2.3.4`
- Wrong: `v0.1.1`, `0.1`, `release-0.1.1`

## What Happens Automatically

When you push a correctly formatted tag, GitHub Actions will:

1. Check out the code
2. Build the Windows executable
3. Extract the changelog for this version
4. Create a GitHub Release with:
   - Release name: `vX.Y.Z`
   - Release notes from CHANGELOG.md
   - Attached binary: `claude-code-switcher.exe`

## Verifying the Release

After pushing the tag:

1. Check the Actions tab: https://github.com/fanis/claude-code-switcher/actions
2. Verify the release was created: https://github.com/fanis/claude-code-switcher/releases
