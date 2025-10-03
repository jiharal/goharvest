# GoHarvest - OAI-PMH Harvesting Library

[![Go Reference](https://pkg.go.dev/badge/github.com/jiharal/goharvest.svg)](https://pkg.go.dev/github.com/jiharal/goharvest)
[![Go Report Card](https://goreportcard.com/badge/github.com/jiharal/goharvest)](https://goreportcard.com/report/github.com/jiharal/goharvest)

GoHarvest adalah library Go untuk harvesting metadata dari OAI-PMH repositories. Library ini menyediakan **unified API** yang clean dan type-safe untuk berbagai format metadata (MARCXML, Dublin Core, dll) melalui satu entry point.

## Features

- ✅ **Unified API** - Single function untuk semua metadata formats
- ✅ **Type-Safe** - Interface-based design dengan proper type assertions
- ✅ **Zero Duplication** - Clean, maintainable codebase
- ✅ **Resumption Token Support** - Automatic pagination handling
- ✅ **Format Agnostic** - Easy to extend untuk format baru
- ✅ **Backward Compatible** - Legacy APIs tetap didukung
- ✅ **Production Ready** - Tested dengan real-world OAI-PMH endpoints

## Installation

```bash
go get github.com/jiharal/goharvest
```

## Quick Start - Unified API

### Single Entry Point untuk Semua Format

```go
package main

import (
    "fmt"
    "log"
    "github.com/jiharal/goharvest"
)

func main() {
    client := goharvest.NewClient("https://balaiyanpus.jogjaprov.go.id/opac/oai")

    // Unified Harvest - otomatis detect format
    err := client.Harvest("marcxml", func(response goharvest.OAIResponse) error {
        records := response.GetRecords()

        for _, record := range records {
            // Extract metadata
            metadata := record.ExtractMetadata()

            // Type assertion based on format
            if bookMeta, ok := metadata.(*goharvest.BookMetadata); ok {
                fmt.Printf("Title: %s\n", bookMeta.Title)
                fmt.Printf("Author: %s\n", bookMeta.MainAuthor)
            }
        }

        return nil
    })

    if err != nil {
        log.Fatal(err)
    }
}
```

## Keuntungan Unified API

### 1. **Single Function Call**
Service lain hanya perlu memanggil `client.Harvest()` untuk semua format metadata.

### 2. **Format Agnostic**
Tidak perlu tahu detail format metadata di awal. Framework otomatis menangani parsing.

### 3. **Type-Safe dengan Interface**
```go
// Interface MetadataExtractor
type MetadataExtractor interface {
    ExtractMetadata() interface{}
    GetFormat() MetadataFormat
}

// Interface OAIResponse
type OAIResponse interface {
    GetRecords() []MetadataExtractor
    GetResumptionToken() string
    HasError() bool
    GetError() *OAIError
}
```

## Examples

### Example 1: Harvest MARCXML

```go
client := goharvest.NewClient("https://balaiyanpus.jogjaprov.go.id/opac/oai")

err := client.Harvest("marcxml", func(response goharvest.OAIResponse) error {
    records := response.GetRecords()

    for _, record := range records {
        metadata := record.ExtractMetadata()

        if bookMeta, ok := metadata.(*goharvest.BookMetadata); ok {
            fmt.Printf("ISBN: %s\n", bookMeta.ISBN)
            fmt.Printf("Title: %s\n", bookMeta.Title)
            fmt.Printf("Publisher: %s\n", bookMeta.Publisher)
        }
    }

    return nil
})
```

### Example 2: Harvest Dublin Core

```go
client := goharvest.NewClient("https://example.com/oai")

err := client.Harvest("oai_dc", func(response goharvest.OAIResponse) error {
    records := response.GetRecords()

    for _, record := range records {
        metadata := record.ExtractMetadata()

        if dcMeta, ok := metadata.(*goharvest.DCMetadata); ok {
            fmt.Printf("Title: %v\n", dcMeta.Title)
            fmt.Printf("Creator: %v\n", dcMeta.Creator)
            fmt.Printf("Subject: %v\n", dcMeta.Subject)
        }
    }

    return nil
})
```

### Example 3: Generic Handler untuk Multiple Formats

```go
func handleMetadata(response goharvest.OAIResponse) error {
    records := response.GetRecords()

    for _, record := range records {
        format := record.GetFormat()

        switch format {
        case goharvest.FormatMARCXML:
            metadata := record.ExtractMetadata()
            if bookMeta, ok := metadata.(*goharvest.BookMetadata); ok {
                // Handle MARCXML
                processBook(bookMeta)
            }

        case goharvest.FormatOAIDC:
            metadata := record.ExtractMetadata()
            if dcMeta, ok := metadata.(*goharvest.DCMetadata); ok {
                // Handle Dublin Core
                processDublinCore(dcMeta)
            }
        }
    }

    return nil
}

client.Harvest("marcxml", handleMetadata)
client.Harvest("oai_dc", handleMetadata)
```

## Backward Compatibility

API lama tetap berfungsi untuk memastikan kompatibilitas:

### Old MARCXML API (masih didukung)
```go
err := client.HarvestAll("marcxml", func(o *goharvest.OAIPMHResponse) error {
    metadatas := o.ExtractAllBookMetadata()
    // ... process metadatas
    return nil
})
```

### Old Dublin Core API (masih didukung)
```go
err := client.HarvestAllDC("oai_dc", func(o *goharvest.OAIPMHResponseDC) error {
    metadatas := o.ExtractAllDCMetadata()
    // ... process metadatas
    return nil
})
```

## Architecture

GoHarvest menggunakan **clean architecture** dengan separation of concerns:

```
┌─────────────────────────────────────────────────┐
│           Service / Application Layer           │
│                                                 │
│    Single Call: client.Harvest("format", cb)   │
└─────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────┐
│              harvester.go (Unified)             │
│                                                 │
│  - Harvest() - Main entry point                │
│  - Format detection & routing                  │
│  - harvestWithParser() - Unified harvest loop  │
│  - performListRecordsRequest() - HTTP layer    │
└─────────────────────────────────────────────────┘
                        │
        ┌───────────────┴───────────────┐
        ▼                               ▼
┌──────────────────┐          ┌──────────────────┐
│  marchxml.go     │          │    oai_dc.go     │
│                  │          │                  │
│  MARCRecord      │          │  DublinCore      │
│  BookMetadata    │          │  DCMetadata      │
│                  │          │                  │
│  implements:     │          │  implements:     │
│  - OAIResponse   │          │  - OAIResponse   │
│  - Metadata      │          │  - Metadata      │
│    Extractor     │          │    Extractor     │
└──────────────────┘          └──────────────────┘
```

### Key Design Decisions

1. **Unified HTTP Layer** - Single `performListRecordsRequest()` untuk semua formats
2. **Generic Harvest Loop** - `harvestWithParser()` eliminates code duplication
3. **Interface-Based** - Type-safe dengan runtime flexibility
4. **Zero Duplication** - ~200+ lines duplicate code removed

## Supported Metadata Formats

| Format | Constant | Description |
|--------|----------|-------------|
| MARCXML | `FormatMARCXML` | Machine-Readable Cataloging XML |
| Dublin Core | `FormatOAIDC` | OAI Dublin Core |

## Error Handling

```go
err := client.Harvest("marcxml", func(response goharvest.OAIResponse) error {
    // Check for OAI-PMH errors
    if response.HasError() {
        err := response.GetError()
        return fmt.Errorf("OAI error [%s]: %s", err.Code, err.Message)
    }

    // Process records...

    return nil
})

if err != nil {
    log.Printf("Harvest failed: %v", err)
}
```

## API Reference

### Core Types

```go
// OAIClient - Main client for OAI-PMH operations
type OAIClient struct {
    BaseURL    string
    HTTPClient *http.Client
}

// OAIResponse - Unified interface for all response types
type OAIResponse interface {
    GetRecords() []MetadataExtractor
    GetResumptionToken() string
    HasError() bool
    GetError() *OAIError
}

// MetadataExtractor - Interface for metadata extraction
type MetadataExtractor interface {
    ExtractMetadata() interface{}
    GetFormat() MetadataFormat
}
```

### Core Functions

```go
// NewClient creates a new OAI-PMH client
func NewClient(baseURL string) *OAIClient

// Harvest - Unified API (Recommended)
func (c *OAIClient) Harvest(metadataPrefix string, callback HarvestCallback) error

// HarvestAll - Legacy MARCXML API (Backward Compatible)
func (c *OAIClient) HarvestAll(metadataPrefix string, callback func(*OAIPMHResponse) error) error

// HarvestAllDC - Legacy Dublin Core API (Backward Compatible)
func (c *OAIClient) HarvestAllDC(metadataPrefix string, callback func(*OAIPMHResponseDC) error) error
```

## Testing

Run all tests:
```bash
go test -v
```

Run specific test:
```bash
go test -v -run TestUnifiedHarvestMARCXML
```

Run tests with coverage:
```bash
go test -cover -v
```

### Test Examples

The library includes comprehensive tests for:
- ✅ MARCXML parsing and extraction
- ✅ Dublin Core parsing and extraction
- ✅ Unified API with multiple formats
- ✅ Backward compatibility
- ✅ Real-world endpoints (UAD, UGM, UNY, UTDI, AMIKOM)

See `harvester_test.go`, `marchxml_test.go`, and `oai_dc_test.go` for examples.

## Migration Guide

### Dari API Lama ke Unified API

**Before:**
```go
// Separate calls untuk setiap format
client.HarvestAll("marcxml", marcxmlCallback)
client.HarvestAllDC("oai_dc", dcCallback)
```

**After:**
```go
// Single unified call
client.Harvest("marcxml", unifiedCallback)
client.Harvest("oai_dc", unifiedCallback)
```

**Benefits:**
- ✅ Konsisten API
- ✅ Easier testing
- ✅ Extensible untuk format baru
- ✅ Backward compatible

## Performance & Best Practices

### Resumption Token Handling

GoHarvest automatically handles pagination via resumption tokens:

```go
// Automatic pagination - no manual token management needed
client.Harvest("marcxml", func(response goharvest.OAIResponse) error {
    // Process each batch
    records := response.GetRecords()
    fmt.Printf("Processing %d records\n", len(records))
    return nil
})
```

### Error Handling Best Practices

```go
client.Harvest("marcxml", func(response goharvest.OAIResponse) error {
    // 1. Check OAI-PMH level errors
    if response.HasError() {
        err := response.GetError()
        return fmt.Errorf("OAI error [%s]: %s", err.Code, err.Message)
    }

    // 2. Process records with type checking
    for _, record := range response.GetRecords() {
        metadata := record.ExtractMetadata()

        if bookMeta, ok := metadata.(*goharvest.BookMetadata); ok {
            // Safe to use bookMeta
            processBook(bookMeta)
        } else {
            log.Printf("Unexpected metadata type: %T", metadata)
        }
    }

    return nil
})
```

### Batch Processing

For large datasets, process in batches:

```go
totalRecords := 0
client.Harvest("marcxml", func(response goharvest.OAIResponse) error {
    records := response.GetRecords()
    totalRecords += len(records)

    // Process batch
    for _, record := range records {
        // Your processing logic
    }

    fmt.Printf("Processed %d total records\n", totalRecords)
    return nil
})
```

## Contributing

Contributions are welcome! Untuk menambahkan metadata format baru:

1. Implementasikan `MetadataExtractor` interface untuk metadata type baru
2. Implementasikan `OAIResponse` interface untuk response type
3. Tambahkan parser function di `harvester.go`
4. Tambahkan case di `Harvest()` switch statement
5. Tambahkan constant di `metadata.go`
6. Tambahkan tests

### Development Setup

```bash
# Clone repository
git clone https://github.com/jiharal/goharvest.git
cd goharvest

# Run tests
go test -v

# Build
go build ./...
```

## Real-World Usage

GoHarvest telah digunakan untuk harvesting dari berbagai repository:

- 🎓 **UAD EPrints** - https://eprints.uad.ac.id/cgi/oai2
- 📚 **UGM Lexicon** - https://jurnal.ugm.ac.id/lexicon/oai
- 🏛️ **UNY EPrints** - https://eprints.uny.ac.id/cgi/oai2
- 📖 **UTDI EPrints** - https://eprints.utdi.ac.id/cgi/oai2
- 📰 **AMIKOM Journal** - https://jurnal.amikom.ac.id/index.php/joism/oai
- 🏢 **Balai Yanpus Yogyakarta** - https://balaiyanpus.jogjaprov.go.id/opac/oai

## Changelog

### v1.0.0 (Latest)
- ✅ Unified API implementation
- ✅ Removed ~200+ lines of duplicate code
- ✅ Generic harvest loop with `harvestWithParser()`
- ✅ Interface-based design
- ✅ Backward compatible APIs
- ✅ Comprehensive test coverage

## License

MIT License - see LICENSE file for details

## Authors

- [@jiharal](https://github.com/jiharal)

## Acknowledgments

Special thanks to all OAI-PMH repository maintainers yang menyediakan public endpoints untuk testing dan development.
