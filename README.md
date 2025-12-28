# Google Docs Manager

A comprehensive command-line tool for managing Google Docs operations, including creating, reading, formatting, tables, images, and more.

## Features

- **Document Operations**: Create, copy, read, and get document information
- **Content Management**: Set content from markdown, update sections, insert text
- **Formatting**: Bold, italic, underline, colors, font sizes, paragraph alignment
- **Lists**: Create bulleted and numbered lists, remove list formatting
- **Tables**: Insert tables, update cell content, style cells with background colors
- **Images**: Insert images with optional size specifications
- **Structure**: Add headers and footers, get document structure
- **Markdown Support**: Convert between Google Docs and Markdown formats

## Installation

### Prerequisites

- Go 1.25.4 or later
- Google Cloud Project with Docs API and Drive API enabled
- OAuth 2.0 credentials file

### Building from Source

```bash
# Clone the repository
git clone <repository-url>
cd google-docs-manager

# Build the binary
make build

# Install to /usr/local/bin
make install

# Or install to a custom location
TARGET=/custom/path make install
```

## Setup

### 1. Create Google Cloud Project Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google Docs API and Google Drive API
4. Create OAuth 2.0 credentials:
   - Go to "Credentials" → "Create Credentials" → "OAuth client ID"
   - Choose "Desktop app" as the application type
   - Download the credentials file

### 2. Configure Credentials

Place your credentials file in `~/.gdrive/credentials.json`:

```bash
mkdir -p ~/.gdrive
mv ~/Downloads/credentials.json ~/.gdrive/
```

On first run, the tool will prompt you to authorize access and save a token file in `~/.gdrive/token.json`.

## Usage

### Document Operations

```bash
# Create a new document
google-docs-manager create "My Document"

# Create a document in a specific folder
google-docs-manager create "My Document" --folder <folder-id>

# Copy an existing document
google-docs-manager copy <source-doc-id> "New Document Title"

# Read a document as markdown
google-docs-manager read <document-id>

# Get document information
google-docs-manager info <document-id>

# Get document structure (headings)
google-docs-manager get-structure <document-id>
```

### Content Management

```bash
# Set document content from markdown file
google-docs-manager set-markdown <document-id> content.md

# Update a specific section
google-docs-manager update-section <document-id> "Section Name" content.md

# Insert text after a section
google-docs-manager insert-after <document-id> "Section Name" "Text to insert"

# Delete text in a range
google-docs-manager delete-text <document-id> <start-index> <end-index>
```

### Formatting

```bash
# Format text (bold, italic, underline)
google-docs-manager format-text <document-id> <start-index> <end-index> --bold --italic

# Set text color
google-docs-manager format-text <document-id> <start-index> <end-index> --color "#FF0000"

# Set font size
google-docs-manager format-text <document-id> <start-index> <end-index> --size 14

# Align paragraph
google-docs-manager align-paragraph <document-id> <start-index> <end-index> CENTER

# Create bulleted list
google-docs-manager create-bullets <document-id> <start-index> <end-index>

# Create numbered list
google-docs-manager create-numbered <document-id> <start-index> <end-index>

# Remove list formatting
google-docs-manager remove-bullets <document-id> <start-index> <end-index>
```

### Tables

```bash
# Insert a 3x4 table at index 1
google-docs-manager insert-table <document-id> 1 3 4

# Update table cell content
google-docs-manager update-table-cell <document-id> <table-start-index> <row> <col> "Cell Text"

# Style table cell background
google-docs-manager style-table-cell <document-id> <table-start-index> <row> <col> --bg-color "#FFFF00"
```

### Images

```bash
# Insert an image
google-docs-manager insert-image <document-id> <index> <image-url>

# Insert an image with specific dimensions
google-docs-manager insert-image <document-id> <index> <image-url> --width 400 --height 300
```

### Structure

```bash
# Add header
google-docs-manager add-header <document-id> "Header Text"

# Add footer
google-docs-manager add-footer <document-id> "Footer Text"
```

## Project Structure

```
google-docs-manager/
├── cmd/
│   └── google-docs-manager/    # Main application entry point
│       └── main.go
├── internal/                    # Private application code
│   ├── auth/                   # OAuth authentication
│   ├── cli/                    # CLI commands
│   ├── conversion/             # Markdown ↔ Docs conversion
│   └── document/               # Document operations
├── Makefile                    # Build automation
├── go.mod                      # Go module definition
├── go.sum                      # Dependency checksums
├── README.md                   # This file
└── CLAUDE.md                   # AI-oriented documentation
```

## Development

### Available Make Targets

```bash
make build      # Build the binary
make rebuild    # Clean and rebuild from scratch
make install    # Build and install to /usr/local/bin
make uninstall  # Remove installed binary
make clean      # Remove build artifacts
make test       # Run tests
make fmt        # Format code
make vet        # Run go vet
make check      # Run fmt, vet, and test
make help       # Show help message
```

### Code Formatting

The project follows standard Go conventions:

```bash
# Format all code
make fmt

# Check for issues
make vet

# Run all checks
make check
```

## Architecture

The project follows the Standard Go Project Layout:

- **cmd/**: Entry points for the application (minimal logic)
- **internal/**: Private application code organized by domain
  - **auth**: OAuth2 authentication with Google APIs
  - **cli**: Cobra-based CLI commands
  - **conversion**: Markdown and color conversion utilities
  - **document**: Document structure operations

All business logic is in `internal/` packages. The entry point in `cmd/google-docs-manager/main.go` handles only initialization and wiring.

## Error Handling

All errors are properly wrapped with context using `%w` for error chains. Error messages include:
- Technical details for debugging
- Actionable guidance for users
- Proper error propagation up the call stack

## License

[Your License Here]

## Contributing

[Your Contributing Guidelines Here]
