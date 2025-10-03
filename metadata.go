package goharvest

// MetadataFormat represents the type of metadata format
type MetadataFormat string

const (
	FormatMARCXML MetadataFormat = "marcxml"
	FormatOAIDC   MetadataFormat = "oai_dc"
)

// MetadataExtractor is the interface for all metadata extractors
type MetadataExtractor interface {
	// ExtractMetadata extracts metadata from the record
	ExtractMetadata() interface{}
	// GetFormat returns the metadata format type
	GetFormat() MetadataFormat
}

// OAIResponse is the unified interface for all OAI-PMH responses
type OAIResponse interface {
	// GetRecords returns all records in the response
	GetRecords() []MetadataExtractor
	// GetResumptionToken returns the resumption token if available
	GetResumptionToken() string
	// HasError returns true if the response contains an error
	HasError() bool
	// GetError returns the error information
	GetError() *OAIError
}

// Common OAI-PMH structures are defined in marchxml.go and oai_dc.go
// We reference them here through the interfaces

// HarvestCallback is the callback function type for harvest operations
type HarvestCallback func(response OAIResponse) error

// DateRange represents the date range filter for selective harvesting
// Dates should be in UTC and formatted as YYYY-MM-DD or YYYY-MM-DDThh:mm:ssZ
type DateRange struct {
	// From specifies the lower bound (inclusive) for datestamp-based selective harvesting
	From string
	// Until specifies the upper bound (inclusive) for datestamp-based selective harvesting
	Until string
}
