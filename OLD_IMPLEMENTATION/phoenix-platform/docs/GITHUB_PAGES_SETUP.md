# GitHub Pages Setup Guide

## Overview
This guide walks through setting up GitHub Pages to host the Phoenix Platform documentation.

## Prerequisites
- Repository admin access
- GitHub Pages enabled for the repository

## Setup Steps

### 1. Enable GitHub Pages
1. Go to repository Settings → Pages
2. Under "Build and deployment":
   - Source: GitHub Actions
   - ✅ This is already configured in `.github/workflows/docs.yml`

### 2. Repository Settings
No additional configuration needed - the workflow handles everything:
- Builds on push to main branch
- Deploys to GitHub Pages automatically
- Supports versioning with mike

### 3. First Deployment
After merging the PR:
1. The workflow will trigger automatically
2. Monitor progress in Actions tab
3. Once complete, docs will be available at:
   ```
   https://deepaucksharma-nr.github.io/Phoenix/
   ```

### 4. Custom Domain (Optional)
To use a custom domain:
1. Go to Settings → Pages
2. Add custom domain
3. Update DNS records as instructed

## Workflow Features

### Automatic Deployment
- Triggers on push to main branch
- Only rebuilds when docs change
- Deploys to GitHub Pages environment

### Version Management
- Uses `mike` for versioned documentation
- Automatically tags releases
- Maintains "latest" alias

### Build Validation
- Strict mode ensures no broken links
- Validates all markdown syntax
- Checks for missing includes

## Local Testing

### Build Documentation
```bash
cd phoenix-platform
mkdocs build --strict
```

### Serve Locally
```bash
cd phoenix-platform
mkdocs serve
# Visit http://localhost:8000
```

### Test Versioning
```bash
# Install mike
pip install mike

# Deploy version
mike deploy 0.1.0 latest --update-aliases

# List versions
mike list

# Serve with versions
mike serve
```

## Troubleshooting

### Build Failures
1. Check Actions tab for error logs
2. Common issues:
   - Missing requirements: Update `docs/requirements.txt`
   - Broken links: Fix in markdown files
   - Invalid YAML: Check `mkdocs.yml` syntax

### Pages Not Updating
1. Verify workflow completed successfully
2. Check Pages settings are correct
3. Clear browser cache
4. Wait 5-10 minutes for CDN propagation

### Permission Errors
Ensure workflow has permissions:
```yaml
permissions:
  contents: write
  pages: write
  id-token: write
```

## Maintenance

### Update Dependencies
```bash
# Update requirements
cd phoenix-platform
pip install --upgrade mkdocs mkdocs-material mike
pip freeze > docs/requirements.txt
```

### Monitor Usage
- Check Pages settings for bandwidth usage
- Review deployment frequency
- Monitor build times in Actions

## Next Steps
1. Merge PR to trigger first deployment
2. Verify documentation is accessible
3. Set up custom domain if needed
4. Add documentation badge to README:
   ```markdown
   [![Documentation](https://img.shields.io/badge/docs-GitHub%20Pages-blue)](https://deepaucksharma-nr.github.io/Phoenix/)
   ```