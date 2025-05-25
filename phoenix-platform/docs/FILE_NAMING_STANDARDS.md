# File Naming Standards

## Documentation Files (.md)

### Standard Convention: UPPERCASE_WITH_UNDERSCORES

All markdown documentation files should follow these rules:

1. **Use UPPERCASE letters**: `README.md`, `TESTING.md`
2. **Use underscores for spaces**: `QUICK_START_GUIDE.md`
3. **Use descriptive names**: `TECHNICAL_SPEC_API_SERVICE.md`
4. **Exception**: Standard files that tools expect in lowercase:
   - `README.md` (GitHub standard)
   - Generated API docs if required by tools

### Current Files to Rename

| Current Name | New Name | Location |
|--------------|----------|----------|
| `api-reference.md` | `API_REFERENCE.md` | `phoenix-platform/docs/` |
| `architecture.md` | `ARCHITECTURE.md` | `phoenix-platform/docs/` |
| `pipeline-guide.md` | `PIPELINE_GUIDE.md` | `phoenix-platform/docs/` |
| `troubleshooting.md` | `TROUBLESHOOTING.md` | `phoenix-platform/docs/` |
| `user-guide.md` | `USER_GUIDE.md` | `phoenix-platform/docs/` |
| `examples.md` | `EXAMPLES.md` | `phoenix-platform/pkg/interfaces/` |

### Directory Structure

Directories should remain lowercase with hyphens:
- `docs/`
- `planning/`
- `reviews/`
- `technical-specs/`

### Benefits

1. **Consistency**: All docs follow same pattern
2. **Visibility**: Uppercase files stand out in listings
3. **Clarity**: Clear distinction between docs and code files
4. **Tradition**: Follows common open source practices