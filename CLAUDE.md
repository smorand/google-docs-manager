# Google Docs Manager - AI Documentation

## Project Overview

**Purpose**: A comprehensive CLI tool for managing Google Docs operations programmatically.

**Technology Stack**:
- Go 1.25.4
- Cobra (CLI framework)
- Google Docs API v1
- Google Drive API v3
- OAuth2 authentication

**Project Type**: CLI application following Standard Go Project Layout

## Architecture

### Directory Structure

```
google-docs-manager/
├── cmd/google-docs-manager/     # Entry point (main package)
│   └── main.go                  # Minimal initialization only
├── internal/                     # Private application code
│   ├── auth/                    # OAuth2 authentication
│   │   └── auth.go              # Client creation, token management
│   ├── cli/                     # CLI commands (Cobra)
│   │   ├── root.go              # Root command and initialization
│   │   ├── document.go          # Document CRUD operations
│   │   ├── content.go           # Content management
│   │   ├── formatting.go        # Text and paragraph formatting
│   │   ├── table.go             # Table operations
│   │   ├── image.go             # Image operations
│   │   └── structure.go         # Headers/footers
│   ├── conversion/              # Conversion utilities
│   │   ├── markdown.go          # Markdown ↔ Docs conversion
│   │   └── colors.go            # Hex color → RGB conversion
│   └── document/                # Document domain logic
│       └── structure.go         # Section/heading operations
├── go.mod                       # Module definition
├── go.sum                       # Dependency checksums
├── Makefile                     # Build automation
├── README.md                    # User documentation
└── CLAUDE.md                    # This file (AI documentation)
```

### Package Responsibilities

#### `cmd/google-docs-manager`
- **Purpose**: Application entry point only
- **Key Files**: `main.go`
- **Responsibilities**: Call `cli.Execute()` and handle exit
- **Note**: Contains NO business logic

#### `internal/auth`
- **Purpose**: OAuth2 authentication with Google APIs
- **Key Functions**:
  - `GetClient(ctx)`: Returns authenticated HTTP client
  - `GetDocsService(ctx)`: Returns Google Docs API service
  - `GetDriveService(ctx)`: Returns Google Drive API service
- **Credentials**: Stored in `~/.gdrive/credentials.json` and `~/.gdrive/token.json`

#### `internal/cli`
- **Purpose**: CLI command definitions and handlers
- **Framework**: spf13/cobra
- **Pattern**: Each command has:
  - Command definition (cobra.Command)
  - Handler function (runXxx)
  - Flag initialization (initXxxCommands)
- **Color Output**: Uses fatih/color for green (success), red (error), cyan (info)

#### `internal/conversion`
- **Purpose**: Data conversion utilities
- **Key Functions**:
  - `DocsToMarkdown(doc)`: Convert Google Doc to Markdown
  - `MarkdownToDocsRequests(markdown, startIndex)`: Convert Markdown to API requests
  - `ParseColor(hexColor)`: Convert hex color to Google Docs RGB format
  - `GetParagraphText(paragraph)`: Extract text from paragraph

#### `internal/document`
- **Purpose**: Document structure operations
- **Key Types**:
  - `Section`: Represents heading with title, level, start/end indices
- **Key Functions**:
  - `GetStructure(doc)`: Extract all sections from document
  - `FindSection(doc, name)`: Find section by name

## Key Patterns

### Command Structure
All CLI commands follow this pattern:
```go
var commandCmd = &cobra.Command{
    Use:   "command <args>",
    Short: "Description",
    Args:  cobra.ExactArgs(n),
    RunE:  runCommand,
}

func runCommand(cmd *cobra.Command, args []string) error {
    ctx := context.Background()
    // 1. Parse arguments
    // 2. Get service via auth package
    // 3. Execute API operations
    // 4. Output results (JSON to stdout, status to stderr)
    return nil
}
```

### Error Handling
- All errors use `%w` for wrapping to maintain error chains
- Context is added when propagating errors up the stack
- User-facing errors include actionable guidance
- Technical errors are wrapped with meaningful context

### API Request Pattern
Most operations follow:
1. Get authenticated service
2. Get current document state (if needed)
3. Build request(s) as `[]*docs.Request`
4. Execute `BatchUpdate` with requests
5. Output success message to stderr, data to stdout

### Output Convention
- **stdout**: Machine-readable output (JSON, IDs, markdown)
- **stderr**: Human-readable status messages (colored)
- Success messages use green color
- Info messages use cyan color

## Common Operations

### Adding New Commands

1. Create command definition in appropriate CLI file
2. Add handler function with signature: `func runXxx(cmd *cobra.Command, args []string) error`
3. Add flags in `initXxxCommands()` if needed
4. Register command in `cli/root.go` `initCommands()`

### Working with Document Indices

Google Docs uses character-based indices:
- Document starts at index 1 (index 0 is implicit)
- Ranges are `[startIndex, endIndex)` (exclusive end)
- Getting document gives end index: `doc.Body.Content[len(doc.Body.Content)-1].EndIndex`
- Use UTF-8 rune counting, not byte counting: `utf8.RuneCountInString(text)`

### Markdown Conversion

The conversion supports:
- Headings (`#` through `######` → HEADING_1 through HEADING_6)
- Bold (`**text**` or `__text__`)
- Italic (`*text*` or `_text_`)
- Tables (basic markdown table format)
- Links (`[text](url)`)

### Authentication Flow

1. Check for `~/.gdrive/credentials.json` (OAuth client credentials)
2. Try to load `~/.gdrive/token.json` (access token)
3. If token missing/expired, prompt user for authorization
4. Save new token for future use
5. Return authenticated HTTP client

## Building and Testing

### Build Commands
```bash
make build      # Build binary
make rebuild    # Clean and rebuild
make install    # Install to /usr/local/bin
make clean      # Remove binary
```

### Code Quality
```bash
make fmt        # Format code
make vet        # Run go vet
make test       # Run tests
make check      # All quality checks
```

### Dependencies

Core dependencies:
- `github.com/spf13/cobra`: CLI framework
- `github.com/fatih/color`: Terminal colors
- `golang.org/x/oauth2`: OAuth2 authentication
- `google.golang.org/api`: Google APIs client library

## Important Notes

### DO NOT Use `/src` Directory
This project previously used `/src` which is discouraged in Go. The new structure follows Standard Go Project Layout with `cmd/` and `internal/`.

### Module Name
Module is `google-docs-manager` (matches binary name)

### Import Paths
All internal imports use full path:
```go
import "google-docs-manager/internal/auth"
import "google-docs-manager/internal/cli"
```

### Alphabetical Ordering
- Struct fields are alphabetically ordered
- Commands in root.go are alphabetically ordered
- Imports are grouped (stdlib, external, internal)

### Context Usage
- Always pass `context.Context` as first parameter
- Create context in command handlers: `ctx := context.Background()`
- Pass context to all service methods

## Future Enhancements

### Header/Footer Text Insertion
Currently, `add-header` and `add-footer` create the structure but don't insert text. This requires:
1. Create header/footer (returns ID in response)
2. Use returned ID to get header/footer section location
3. Insert text at that location in separate BatchUpdate

### Testing
- Add unit tests for conversion functions
- Add integration tests for CLI commands
- Mock Google API responses

### Additional Features
- Batch operations (multiple documents)
- Document templates
- Advanced markdown features (code blocks, nested lists)
- Document sharing permissions
- Document search and listing

## Troubleshooting

### Build Errors
- Run `go mod tidy` to sync dependencies
- Check Go version (requires 1.25.4+)

### Authentication Errors
- Verify credentials file at `~/.gdrive/credentials.json`
- Delete `~/.gdrive/token.json` to force re-authentication
- Ensure APIs are enabled in Google Cloud Console

### API Errors
- Check document ID is valid
- Verify indices are within document bounds
- Ensure API scopes include both Docs and Drive
