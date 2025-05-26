# Service Consolidation Plan

## Analysis Summary

After analyzing both `services/` and `projects/` directories, I've identified the following duplicate services and their state:

### Duplicate Services Analysis

| Service | services/ Status | projects/ Status | Recommendation |
|---------|-----------------|------------------|----------------|
| **analytics** | ✓ Complete implementation<br>✓ go.mod present<br>✗ No Makefile<br>✓ README.md | ✓ Complete implementation<br>✓ go.mod present<br>✓ Makefile<br>✓ README.md<br>✓ VERSION file | **Keep projects/** - More complete structure |
| **anomaly-detector** | ✓ Basic implementation<br>✓ go.mod<br>✗ No Makefile<br>✗ No README | ✓ Structure only<br>✓ go.mod<br>✓ Makefile<br>✓ README.md<br>✓ VERSION file | **Keep services/** - Has actual implementation |
| **api** | ✓ Implementation<br>✓ go.mod<br>✓ Makefile<br>✓ README | ✗ No go.mod<br>✓ Makefile<br>✓ README<br>✓ Structure only | **Keep services/** - Has working code |
| **benchmark** | ✓ Complete implementation<br>✓ go.mod<br>✗ No Makefile | ✓ Complete implementation<br>✓ go.mod<br>✓ Makefile<br>✓ README<br>✓ VERSION | **Keep projects/** - More complete structure |
| **controller** | ✓ Full implementation<br>✓ go.mod<br>✗ No Makefile | ✓ Full implementation<br>✓ go.mod<br>✗ No Makefile | **Keep either** - Both are complete |
| **dashboard** | ✓ Full React app<br>✗ No go.mod<br>✗ No Makefile | ✓ Full React app<br>✗ No go.mod<br>✓ Makefile<br>✓ README<br>59K+ files (node_modules) | **Keep projects/** - Has build infrastructure |
| **generator** | ✓ Implementation<br>✓ go.mod<br>✗ No Makefile | ✗ No go.mod<br>✗ Partial structure | **Keep services/** - Has working code |
| **loadsim-operator** | ✓ Basic structure<br>✓ go.mod | ✓ Complete structure<br>✓ go.mod<br>✓ Makefile<br>✓ README | **Keep projects/** - More complete |
| **phoenix-cli** | ✓ Full implementation<br>✓ go.mod<br>✓ Makefile<br>✓ README | ✗ No go.mod<br>✓ Makefile<br>✓ README<br>✓ Structure | **Keep services/** - Has working code |
| **pipeline-operator** | ✓ Implementation<br>✓ go.mod | ✓ Implementation<br>✓ go.mod<br>✓ Makefile<br>✓ README | **Keep projects/** - More complete structure |

### Services Only in services/

- **collector** (Node.js package)
- **control-actuator-go** (Go service)
- **control-plane/** (with actuator and observer subdirs)
- **generators/** (complex and synthetic subdirs)
- **validator** (Go service)

### Services Only in projects/

- **platform-api** (Go service with Makefile)

## Consolidation Plan

### Phase 1: Move Complete Services from services/ to projects/

1. **anomaly-detector** - Move implementation from services/ to projects/
2. **api** - Move to projects/ and add VERSION file
3. **generator** - Move to projects/ and add Makefile, README, VERSION
4. **phoenix-cli** - Move go.mod and implementation to projects/

### Phase 2: Remove Duplicates (Keep projects/ versions)

1. Remove services/analytics (keep projects/analytics)
2. Remove services/benchmark (keep projects/benchmark)
3. Remove services/dashboard (keep projects/dashboard)
4. Remove services/loadsim-operator (keep projects/loadsim-operator)
5. Remove services/pipeline-operator (keep projects/pipeline-operator)

### Phase 3: Move Unique Services

1. Move services/collector → projects/collector
2. Move services/control-actuator-go → projects/control-actuator-go
3. Move services/validator → projects/validator
4. Handle control-plane services:
   - Move services/control-plane/actuator → projects/control-actuator
   - Move services/control-plane/observer → projects/control-observer
5. Handle generators:
   - Move services/generators/complex → projects/generator-complex
   - Move services/generators/synthetic → projects/generator-synthetic

### Phase 4: Clean up and Update References

1. Update go.work file to remove services/ references
2. Update any import paths in the code
3. Update docker-compose files
4. Update Kubernetes manifests
5. Update CI/CD pipelines
6. Remove empty services/ directory

## Implementation Script

```bash
#!/bin/bash
# consolidate-services.sh

# Phase 1: Move implementations
echo "Phase 1: Moving implementations..."
cp -r services/anomaly-detector/* projects/anomaly-detector/
cp -r services/api/* projects/api/
mkdir -p projects/generator && cp -r services/generator/* projects/generator/
cp -r services/phoenix-cli/go.* projects/phoenix-cli/
cp -r services/phoenix-cli/internal/* projects/phoenix-cli/internal/
cp -r services/phoenix-cli/cmd/* projects/phoenix-cli/cmd/
cp -r services/phoenix-cli/main.go projects/phoenix-cli/

# Phase 2: Remove duplicates
echo "Phase 2: Removing duplicates..."
rm -rf services/analytics
rm -rf services/benchmark
rm -rf services/dashboard
rm -rf services/loadsim-operator
rm -rf services/pipeline-operator

# Phase 3: Move unique services
echo "Phase 3: Moving unique services..."
mv services/collector projects/
mv services/control-actuator-go projects/
mv services/validator projects/
mkdir -p projects/control-actuator && mv services/control-plane/actuator/* projects/control-actuator/
mkdir -p projects/control-observer && mv services/control-plane/observer/* projects/control-observer/
mkdir -p projects/generator-complex && mv services/generators/complex/* projects/generator-complex/
mkdir -p projects/generator-synthetic && mv services/generators/synthetic/* projects/generator-synthetic/

# Phase 4: Update go.work
echo "Phase 4: Updating go.work..."
# This will need manual editing to remove services/ entries
```

## Post-Consolidation Tasks

1. Run `go work sync` to update workspace
2. Run all tests to ensure nothing broke
3. Update documentation
4. Create PR with changes