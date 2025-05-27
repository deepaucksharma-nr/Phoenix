# Git Hooks for Phoenix Platform

This directory contains Git hooks to maintain code quality and prevent common issues.

## Setup

The repository is already configured to use these hooks. If you need to set them up manually:

```bash
git config core.hooksPath .githooks
```

## Available Hooks

### pre-commit
Prevents committing:
- Files larger than 1MB
- Binary files (.exe, .dll, .so, .dylib, .bin)
- Files in bin/ or build/ directories

## Bypass Hooks (Emergency Only)

If you absolutely need to bypass the hooks:

```bash
git commit --no-verify
```

**Warning**: Only bypass hooks when absolutely necessary and with team approval.

## Troubleshooting

### Large File Error
If you get an error about large files:
1. Check if the file should be in the repository
2. If it's a build artifact, add it to .gitignore
3. If it's a necessary large file, consider using Git LFS
4. Run `make clean-binaries` to remove build artifacts

### Binary File Error
If you get an error about binary files:
1. Binary files should not be committed
2. Add the file pattern to .gitignore
3. Use release systems for distributing binaries

## Adding New Hooks

To add a new hook:
1. Create the hook file in this directory
2. Make it executable: `chmod +x .githooks/hook-name`
3. Document it in this README