# Nori Registry Structure

The nori registry is a GitHub repository that contains package manifests in a simple, declarative YAML format.

## Repository Structure

The registry repository should follow this structure:

```
registry/
├── index.yaml              # Package index listing all available packages
└── packages/
    ├── node.yaml          # Node.js package manifest
    ├── python.yaml        # Python package manifest
    ├── deno.yaml          # Deno package manifest
    └── ...                # Other package manifests
```

## Accessing the Registry

The registry is accessed via GitHub's raw content URLs:

- **Base URL format**: `https://raw.githubusercontent.com/{owner}/{repo}/{branch}`
- **Index URL**: `{baseURL}/index.yaml`
- **Package manifest URL**: `{baseURL}/packages/{package-name}.yaml`

### Example

If your registry is at `https://github.com/chirag-bruno/nori-registry` on the `main` branch:

- Base URL: `https://raw.githubusercontent.com/chirag-bruno/nori-registry/main`
- Index: `https://raw.githubusercontent.com/chirag-bruno/nori-registry/main/index.yaml`
- Neovim manifest: `https://raw.githubusercontent.com/chirag-bruno/nori-registry/main/packages/neovim.yaml`

## Configuration

The registry URL can be configured via the `NORI_REGISTRY_URL` environment variable:

```bash
export NORI_REGISTRY_URL="https://raw.githubusercontent.com/chirag-bruno/nori-registry/main"
```

If not set, nori defaults to: `https://raw.githubusercontent.com/chirag-bruno/nori-registry/main`

## Index Format

The `index.yaml` file lists all available packages:

```yaml
packages:
  - name: node
    description: Node.js runtime
  - name: python
    description: Python programming language
  - name: deno
    description: Deno runtime
```

## Package Manifest Format

Each package has a manifest file in `packages/{name}.yaml`. See [MANIFEST.md](../schema/manifest-v1.schema.json) for the full schema.

Example `packages/node.yaml`:

```yaml
schema: 1
name: node
description: Node.js runtime
homepage: https://nodejs.org
license: MIT
bins:
  - bin/node
  - bin/npm
  - bin/npx
versions:
  "22.2.0":
    platforms:
      linux-amd64:
        type: tar
        url: https://nodejs.org/dist/v22.2.0/node-v22.2.0-linux-x64.tar.xz
        checksum: sha256:5f4a1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab
      darwin-arm64:
        type: tar
        url: https://nodejs.org/dist/v22.2.0/node-v22.2.0-darwin-arm64.tar.gz
        checksum: sha256:9a2c1234567890abcdef1234567890abcdef1234567890abcdef1234567890cd
```

## Creating a Registry

1. Create a new GitHub repository (e.g., `nori-registry`)
2. Create the directory structure:
   ```bash
   mkdir -p packages
   ```
3. Add `index.yaml` at the root
4. Add package manifests in `packages/`
5. Commit and push to GitHub
6. Set `NORI_REGISTRY_URL` to point to your repository's raw content URL

## Testing Your Registry

You can test your registry using the integration test:

```bash
NORI_TEST_REGISTRY_URL=https://raw.githubusercontent.com/your-org/your-registry/main go test ./internal/registry -v -run TestRegistryIntegrationWithGitHub
```

## Best Practices

1. **Use semantic versioning**: Package versions should follow `MAJOR.MINOR.PATCH` format
2. **Validate manifests**: Use the JSON schema to validate manifests before committing
3. **Use HTTPS URLs**: All asset URLs must use HTTPS
4. **Include checksums**: All assets must have SHA256 checksums
5. **Keep manifests updated**: Update manifests when new versions are released
6. **Document packages**: Include clear descriptions in the index

## CI/CD Integration

Consider setting up CI/CD to:
- Validate all manifests against the JSON schema
- Verify asset URLs are accessible
- Check checksums match
- Ensure all platforms are covered for each version

