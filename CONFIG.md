# Configuration Guide

Archi supports flexible configuration through multiple sources with the following priority order:

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Configuration files** (YAML or JSON)
4. **Default values** (lowest priority)

## Configuration File Formats

### YAML Configuration (Recommended)

Create `config.yaml`:

```yaml
# API Configuration
apiBaseURL: "http://localhost:3005"

# Output Configuration
defaultOutputDir: "."
jsonOutputFile: "output.json"
markdownOutputFile: "output.md"
reportOutputFile: "report.md"
estimationFile: "estimation.md"

# AI Model Configuration
fileAnalysisModel: "mistral-small-2501"
folderAnalysisModel: "mistral-small-2501"
architectureAnalysisModel: "mistral-small-2501"
imageAnalysisModel: "magistral-small-2509"

# Processing Configuration
maxFileSize: 1048576  # 1MB in bytes
requestDelay: "200ms" # Delay between API requests
```

### JSON Configuration (Legacy Support)

Create `config.json`:

```json
{
  "apiBaseURL": "http://localhost:3005",
  "defaultOutputDir": ".",
  "jsonOutputFile": "output.json",
  "markdownOutputFile": "output.md",
  "reportOutputFile": "report.md",
  "estimationFile": "estimation.md",
  "fileAnalysisModel": "mistral-small-2501",
  "folderAnalysisModel": "mistral-small-2501",
  "architectureAnalysisModel": "mistral-small-2501",
  "imageAnalysisModel": "magistral-small-2509",
  "maxFileSize": 1048576,
  "requestDelay": "200ms"
}
```

## Environment Variables

Override any configuration using environment variables with the `ARCHI_` prefix:

```bash
# API Configuration
export ARCHI_APIBASEURL="http://localhost:3005"

# Output Configuration
export ARCHI_DEFAULTOUTPUTDIR="/tmp/archi-output"
export ARCHI_JSONOUTPUTFILE="analysis.json"
export ARCHI_MARKDOWNOUTPUTFILE="analysis.md"

# Model Configuration
export ARCHI_FILEANALYSISMODEL="mistral-small-2501"
export ARCHI_FOLDERANALYSISMODEL="mistral-small-2501"

# Processing Configuration
export ARCHI_MAXFILESIZE="2097152"  # 2MB
export ARCHI_REQUESTDELAY="300ms"
```

## Configuration File Discovery

Archi searches for configuration files in the following order:

1. **Explicit path**: `--config /path/to/config.yaml`
2. **Current directory**: `./config.yaml` or `./config.json`
3. **Home directory**: `~/.archi/config.yaml`
4. **System directory**: `/etc/archi/config.yaml`

## Parameter Reference

### API Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `apiBaseURL` | string | `http://localhost:3005` | Base URL for the AI API service |

### Output Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `defaultOutputDir` | string | `.` | Directory where output files will be written |
| `jsonOutputFile` | string | `output.json` | Name of the JSON output file |
| `markdownOutputFile` | string | `output.md` | Name of the Markdown output file |
| `reportOutputFile` | string | `report.md` | Name of the architectural analysis report |
| `estimationFile` | string | `estimation.md` | Name of the estimation report file |

### AI Model Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `fileAnalysisModel` | string | `mistral-small-2501` | AI model for file content analysis |
| `folderAnalysisModel` | string | `mistral-small-2501` | AI model for folder analysis |
| `architectureAnalysisModel` | string | `mistral-small-2501` | AI model for architectural analysis |
| `imageAnalysisModel` | string | `magistral-small-2509` | AI model for image analysis |

### Processing Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `maxFileSize` | int | `1048576` | Maximum file size to process (bytes) |
| `requestDelay` | duration | `200ms` | Delay between API requests |

## Usage Examples

### Multiple Configuration Sources

```bash
# Use custom config with environment override
export ARCHI_REQUESTDELAY="500ms"
./archi --config production.yaml /path/to/project
```

### Development vs Production

**Development config** (`dev.yaml`):
```yaml
apiBaseURL: "http://localhost:3005"
requestDelay: "100ms"  # Faster for development
maxFileSize: 2097152   # 2MB for larger files
```

**Production config** (`prod.yaml`):
```yaml
apiBaseURL: "https://api.example.com"
requestDelay: "500ms"  # Slower to be respectful
maxFileSize: 1048576   # 1MB to prevent timeouts
```

### CI/CD Configuration

```bash
# Set via environment in CI
export ARCHI_APIBASEURL="https://internal-ai-service.company.com"
export ARCHI_DEFAULTOUTPUTDIR="/tmp/archi-results"
export ARCHI_REQUESTDELAY="1s"  # Be conservative in CI

./archi count .          # Quick estimation
./archi .               # Full analysis
./archi architecture    # Generate report
```

## Configuration Validation

Archi validates configuration on startup and will report errors for:

- Invalid URLs
- Missing required parameters
- Invalid duration formats
- Negative file sizes
- Inaccessible output directories

Example validation output:
```
Error loading configuration: invalid requestDelay format '200xyz': time: unknown unit "xyz" in duration "200xyz"
```

## Migration from Legacy Configuration

If you have an existing JSON configuration, you can:

1. **Continue using JSON** - No changes needed
2. **Convert to YAML** - Use the examples above
3. **Use both** - YAML takes precedence over JSON

The refactored version is fully backward compatible with existing JSON configurations.