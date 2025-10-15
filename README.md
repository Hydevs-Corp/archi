# Archi - AI-Powered Directory Structure Analyzer

Archi is a powerful Go-based command-line tool that analyzes directory structures and generates AI-powered insights about your project's architecture. It creates detailed reports, visualizations, and architectural recommendations by leveraging AI to understand file contents and folder organization.

## Features

-   🔍 **Smart Directory Analysis**: Recursively scans directories and analyzes file contents
-   🤖 **AI-Powered Insights**: Uses Mistral AI models to understand and describe files and folders
-   📊 **Multiple Output Formats**: Generates JSON, Markdown, and detailed reports
-   🖼️ **Image Analysis**: Supports analysis of images using vision AI
-   📄 **Document Support**: Reads and analyzes DOCX, XLSX, PDF, and text files
-   ⚡ **Estimation Mode**: Quickly estimates processing time before full analysis
-   🏗️ **Architecture Analysis**: Provides detailed architectural recommendations
-   ⚙️ **Modern CLI**: Built with Cobra for intuitive command structure
-   🔧 **Flexible Config**: YAML/JSON configuration with environment variable support
-   🧵 **Batched Requests**: Control concurrency with a configurable batch size

## Roadmap

Here are some of the features and improvements planned for future releases:

-   **Architecture Generation**: Offer the possibility to create a recommended folder architecture on the filesystem.
-   **Multi-AI Provider Support**: Add a way to use other AI providers beyond the current default (e.g., OpenAI, Anthropic).
-   **Enhanced Visualization**: Generate interactive diagrams (e.g., using D3.js or Mermaid.js) of the folder structure and dependencies.
-   **Cost Estimation Improvements**: Refine cost and time estimations based on file types and token counts.
-   **Context Analysis**: Implement a way predefined context to the file, folder, and architecture analysis.
-   **Metadata Injection**: Inject file metadata into the context analysis.
-   **Parallel AI Requests**: Implement parallel processing for AI requests to improve performance.
-   **Advanced File Outlines**: Introduce specific file outlining for security vulnerabilities, and redundant code warnings.
-   **Flat Analysis Mode**: Add a "flat analysis" mode to get a file architecture overview without deep content analysis, while retaining duplicate file warnings.
-   **Expanded Media Support**: Add support for more media file types, including audio and video formats.

## Prerequisites

-   Go 1.25 or later
-   An AI API service running (default: `http://localhost:3005`)
-   The AI service should support:
    -   `/ask` endpoint for text analysis
    -   `/analyze-image` endpoint for image analysis

## Installation

### Building from Source

1. Clone the repository:

```bash
git clone https://github.com/Hydevs-Corp/archi
cd archi
```

2. Install dependencies:

```bash
go mod tidy
```

3. Build the application:

```bash
go build
```

### Using Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/your-username/archi/releases).

#### Windows Binary Usage

1. Download the `archi-windows-amd64.exe` file from the latest release
2. Place it in a directory of your choice (e.g., `C:\tools\archi\`)
3. Optionally, add the directory to your system PATH for global access

**Using the Windows binary:**

```cmd
# Run from the same directory
archi-windows-amd64.exe --help

# If added to PATH, you can use it globally
archi-windows-amd64.exe count C:\your\project

# Analyze current directory
archi-windows-amd64.exe

# Generate architectural recommendations
archi-windows-amd64.exe architecture
```

**Windows Command Examples:**

```cmd
# Quick estimation
archi-windows-amd64.exe count

# Analyze a specific directory
archi-windows-amd64.exe "C:\Users\YourName\Documents\MyProject"

# Generate reports with custom config
archi-windows-amd64.exe --config config.yaml "C:\path\to\project"

# Folders only analysis (faster)
archi-windows-amd64.exe --only-folders "C:\your\project"
```

**Windows Configuration Notes:**

-   Configuration files can be placed in the same directory as the executable
-   Use forward slashes (`/`) or double backslashes (`\\`) in paths within configuration files
-   Environment variables work the same way: `set ARCHI_APIBASEURL=http://localhost:3005`

## Configuration

### YAML Configuration (Recommended)

Create a `config.yaml` file from the example:

```bash
cp config.yaml.example config.yaml
```

Example `config.yaml`:

```yaml
# API Configuration
apiBaseURL: "http://localhost:3005"

# Output Configuration
defaultOutputDir: "."
jsonOutputFile: "output.json"
markdownOutputFile: "output.md"
reportOutputFile: "report.md"
estimationFile: "estimation.md"

# AI Model Configuration (single or multi-model)
# Option A: Single model (string) — only for Mistral when sent as a string
fileAnalysisModel: "mistral-small-2501"
folderAnalysisModel: "mistral-small-2501"
architectureAnalysisModel: "mistral-small-2501"
imageAnalysisModel: "magistral-small-2509"

# Option B: Multi-model (takes precedence over the single-model keys above if non-empty)
# fileAnalysisModels:
#   - provider: "gemini"
#     model: "gemini-2.5-flash"
#   - provider: "mistral"
#     model: "mistral-small-2503"
# folderAnalysisModels: []
# architectureAnalysisModels: []
# imageAnalysisModels: []

# Processing Configuration
maxFileSize: 1048576 # 1MB in bytes
requestDelay: "200ms"
batchSize: 5 # Number of concurrent requests per batch

# Analysis Mode
# Set how analysis runs: "full", "description-only" (no content in JSON), or "folder-only" (folders only)
mode: "full"
```

### JSON Configuration (Legacy)

You can also use JSON configuration:

```bash
cp config.json.example config.json
```

### Configuration Options

**Configuration Parameters:**

-   `apiBaseURL`: Base URL for the AI API service
-   `defaultOutputDir`: Directory where output files will be written
-   `jsonOutputFile`: Name of the JSON output file containing the tree structure
-   `markdownOutputFile`: Name of the Markdown output file with tree visualization
-   `reportOutputFile`: Name of the architectural analysis report file
-   `estimationFile`: Name of the estimation report file (count-only mode)
-   `fileAnalysisModel`: AI model to use for individual file content analysis
-   `folderAnalysisModel`: AI model to use for folder content analysis
-   `architectureAnalysisModel`: AI model to use for architectural analysis
-   `imageAnalysisModel`: AI model to use for image analysis
-   `fileAnalysisModels` / `folderAnalysisModels` / `architectureAnalysisModels` / `imageAnalysisModels`: arrays of `{ provider, model }` entries. When provided, these arrays are sent to the API instead of the single string.
-   `maxFileSize`: Maximum file size to process (in bytes)
-   `requestDelay`: Delay between API requests to avoid overwhelming the service
-   `batchSize`: Number of concurrent requests per batch (default: 5)

### Environment Variables

You can override any configuration using environment variables with the `ARCHI_` prefix:

```bash
export ARCHI_APIBASEURL="http://localhost:3005"
export ARCHI_REQUESTDELAY="300ms"
export ARCHI_MAXFILESIZE="2097152"
```

## Usage

### Command Structure

Archi uses a modern CLI interface with subcommands:

```bash
# Show help
./archi --help

# Show help for specific commands
./archi count --help
./archi architecture --help
```

### Commands

#### Main Analysis Command

```bash
# Analyze current directory
./archi

# Analyze specific directory
./archi /path/to/project

# Analysis modes are configured via config.yaml (mode: full | description-only | folder-only)
# Example: set mode: "folder-only" in config.yaml to only include folders
```

#### Count Command (Quick Estimation)

```bash
# Get quick file/folder count and time estimation
./archi count

# Count specific directory
./archi count /path/to/project
```

#### Architecture Command (Generate Recommendations)

```bash
# Generate architectural recommendations (requires existing analysis)
./archi architecture

# Alternative aliases
./archi arch
./archi archi
```

### Global Flags

-   `--config string`: Path to configuration file (YAML or JSON)
-   `--only-folders`: Only show folders in the output
-   `--no-content`: Exclude file content from the JSON output
-   `--count-only`: Only count files and folders (legacy flag, use `count` command instead)
-   `--better-archi`: Generate architectural recommendations (legacy flag, use `architecture` command instead)

### Usage Examples

1. **Quick estimation** (no AI analysis, fast):

```bash
./archi count
```

2. **Full analysis with custom config**:

```bash
./archi --config my-config.yaml /path/to/project
```

3. **Folders only** (faster, focuses on structure):

```bash
./archi --only-folders
```

4. **Generate architectural report** (run after basic analysis):

```bash
./archi architecture
```

5. **Analysis without file content** (smaller output files):

```bash
./archi --no-content
```

6. **Using environment variables**:

```bash
ARCHI_APIBASEURL="http://custom-api:3005" ./archi
```

7. **Legacy command style** (still supported):

```bash
./archi --count-only     # equivalent to: ./archi count
./archi --better-archi   # equivalent to: ./archi architecture
```

## Output Files

### Generated Files

1. **`output.json`**: Complete directory tree with AI descriptions in JSON format
2. **`output.md`**: Human-readable tree visualization in Markdown
3. **`estimation.md`**: Time estimation report (with `--count-only`)
4. **`report.md`**: Architectural analysis and recommendations (with `--better-archi`)

### File Processing

The tool processes various file types:

-   **Text files**: Content extracted and analyzed
-   **Documents**: DOCX, XLSX, PDF files are parsed
-   **Images**: JPG, PNG, GIF, BMP analyzed with vision AI
-   **Code files**: All programming languages supported
-   **Binary files**: Skipped or analyzed by type

## Workflow Examples

### 1. Quick Project Assessment

```bash
# Get quick overview
./archi count

# Full analysis if time permits
./archi

# Generate architectural recommendations
./archi architecture
```

### 2. Large Project Analysis

```bash
# Start with estimation
./archi count /large/project

# Analyze structure only first
./archi --only-folders /large/project

# Full analysis with content
./archi /large/project

# Generate final report
./archi architecture
```

### 3. Documentation Generation

```bash
# Generate comprehensive documentation
./archi --config doc-config.yaml /project

# Create architectural report
./archi architecture
```

### 4. CI/CD Integration

```bash
# Quick check in CI pipeline
./archi count .

# Generate reports for documentation
./archi --config ci-config.yaml .
./archi architecture

# Upload results to documentation system
```

## Performance Considerations

-   **File processing**: ~4 seconds per file for AI analysis
-   **Folder processing**: ~7 seconds per folder for AI analysis
-   **Request delay**: Configurable delay between API calls (default: 200ms)
-   **Batch size**: Controls concurrency for file and folder analyses (default: 5)
-   **Large projects**: Use `./archi count` first to estimate time
-   **Memory usage**: Large files are truncated to 5000 characters for analysis

## Project Structure

Archi follows Go best practices:

```
├── cmd/                    # CLI commands (Cobra)
│   ├── architecture.go    # Architecture analysis command
│   ├── count.go          # Count estimation command
│   └── root.go           # Root command & CLI setup
├── internal/             # Private application packages
│   ├── analyzer/        # Core analysis logic
│   │   ├── ai_client.go    # AI API communication
│   │   ├── analyzer.go     # Main analysis orchestration
│   │   ├── filereaders.go  # File content extraction
│   │   ├── output.go       # Output generation
│   │   └── types.go        # Core type definitions
│   ├── app/            # Application orchestration
│   │   └── app.go         # High-level app logic
│   └── config/         # Configuration management
│       └── config.go      # Viper-based configuration
├── main.go             # Simple entry point (7 lines)
├── config.yaml.example # YAML configuration template
└── go.mod             # Go module with dependencies
```

## Troubleshooting

### Common Issues

1. **API connection errors**:

    - Ensure AI service is running on configured URL
    - Check firewall settings
    - Verify API endpoints are available

2. **Large file processing**:

    - Files over 5000 characters are truncated for analysis
    - Use `--no-content` to reduce output size
    - Consider `--only-folders` for structure analysis

3. **Permission errors**:

    - Ensure read permissions on analyzed directories
    - Check write permissions for output directory

4. **Memory issues**:
    - Use `./archi count` for very large projects
    - Process subdirectories separately

### Performance Optimization

-   Use custom configuration with appropriate request delays
-   Start with `./archi count` to understand scope
-   Use `--only-folders` for structural analysis
-   Configure output directory to SSD for faster writes

## API Requirements

The tool expects an AI service with these endpoints:

### `/ask` - Text Analysis

```json
POST /ask
{
  "history": [{"role": "user", "content": "..."}],
  "model": "mistral-small-2501"
}
```

### `/analyze-image` - Image Analysis

```json
POST /analyze-image
{
  "image": "base64-encoded-image",
  "model": "mistral-small-2501"
}
```

## CLI Help System

The refactored application includes a comprehensive help system:

```bash
# Main help
./archi --help

# Command-specific help
./archi count --help
./archi architecture --help

# See all available commands
./archi help
```

### Available Commands

-   `count` - Count files and folders and estimate processing time
-   `architecture` (aliases: `arch`, `archi`) - Generate architectural recommendations
-   `completion` - Generate shell completion scripts
-   `help` - Help about any command

## Backwards Compatibility

The refactored version maintains full backwards compatibility:

-   All original flags still work (e.g., `--count-only`, `--better-archi`)
-   Configuration files remain compatible
-   Output formats are unchanged
-   API requirements are identical

New features include:

-   Modern CLI with subcommands
-   YAML configuration support
-   Environment variable configuration
-   Shell completion support
-   Better help system

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

Copyright (c) 2025 Hydevs

## Support

For issues and questions:

-   Check the troubleshooting section
-   Review configuration options
-   Examine output files for error messages
-   Ensure AI service is properly configured
