# How to Create a Release

GitHub Actions will automatically build and release binaries when you push a version tag.

## Create a Release

```bash
# Tag the current commit
git tag v1.0.0

# Push the tag to GitHub
git push origin v1.0.0
```

## What Happens Automatically

The GitHub Actions workflow will:
1. ✅ Build binaries for all platforms:
   - Linux (AMD64, ARM64)
   - macOS (Intel, Apple Silicon)
   - Windows (AMD64)

2. ✅ Create a GitHub Release with:
   - All binaries attached
   - Installation instructions
   - Version information

3. ✅ Users can download binaries directly from the Releases page

## Versioning

Follow semantic versioning:
- `v1.0.0` - Major release
- `v1.1.0` - Minor release (new features)
- `v1.0.1` - Patch release (bug fixes)

## Example Workflow

```bash
# Make changes to your code
git add .
git commit -m "Add new feature"
git push

# Create and push a release tag
git tag v1.1.0
git push origin v1.1.0

# GitHub Actions will automatically build and release!
```

## View Releases

- GoTube: https://github.com/iiTzDante/gotube/releases
- GoMusic: https://github.com/iiTzDante/gomusic/releases
