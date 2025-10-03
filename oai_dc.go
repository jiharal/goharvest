package goharvest

import (
	"encoding/xml"
	"fmt"
)

// DublinCore represents Dublin Core metadata
type DublinCore struct {
	XMLName        xml.Name `xml:"http://www.openarchives.org/OAI/2.0/oai_dc/ dc"`
	SchemaLocation string   `xml:"http://www.w3.org/2001/XMLSchema-instance schemaLocation,attr,omitempty"`
	Title          []string `xml:"http://purl.org/dc/elements/1.1/ title"`
	Creator        []string `xml:"http://purl.org/dc/elements/1.1/ creator"`
	Subject        []string `xml:"http://purl.org/dc/elements/1.1/ subject"`
	Description    []string `xml:"http://purl.org/dc/elements/1.1/ description"`
	Publisher      []string `xml:"http://purl.org/dc/elements/1.1/ publisher"`
	Contributor    []string `xml:"http://purl.org/dc/elements/1.1/ contributor"`
	Date           []string `xml:"http://purl.org/dc/elements/1.1/ date"`
	Type           []string `xml:"http://purl.org/dc/elements/1.1/ type"`
	Format         []string `xml:"http://purl.org/dc/elements/1.1/ format"`
	Identifier     []string `xml:"http://purl.org/dc/elements/1.1/ identifier"`
	Source         []string `xml:"http://purl.org/dc/elements/1.1/ source"`
	Language       []string `xml:"http://purl.org/dc/elements/1.1/ language"`
	Relation       []string `xml:"http://purl.org/dc/elements/1.1/ relation"`
	Coverage       []string `xml:"http://purl.org/dc/elements/1.1/ coverage"`
	Rights         []string `xml:"http://purl.org/dc/elements/1.1/ rights"`
}

// MetadataDC represents the metadata wrapper for Dublin Core
type MetadataDC struct {
	DC  *DublinCore `xml:"http://www.openarchives.org/OAI/2.0/oai_dc/ dc,omitempty"`
	Raw []byte      `xml:",innerxml"`
}

// RecordDC represents an OAI-PMH record with Dublin Core metadata
type RecordDC struct {
	Header   Header     `xml:"header"`
	Metadata MetadataDC `xml:"metadata"`
	About    *About     `xml:"about,omitempty"`
}

// ListRecordsDC contains the list of Dublin Core records from ListRecords verb
type ListRecordsDC struct {
	Records         []RecordDC       `xml:"record"`
	ResumptionToken *ResumptionToken `xml:"resumptionToken,omitempty"`
}

// OAIPMHResponseDC represents the OAI-PMH response with Dublin Core metadata
type OAIPMHResponseDC struct {
	XMLName         xml.Name         `xml:"OAI-PMH"`
	ResponseDate    string           `xml:"responseDate"`
	Request         OAIRequest       `xml:"request"`
	ListRecords     *ListRecordsDC   `xml:"ListRecords,omitempty"`
	GetRecord       *GetRecordDC     `xml:"GetRecord,omitempty"`
	ListIdentifiers *ListIdentifiers `xml:"ListIdentifiers,omitempty"`
	Error           *OAIError        `xml:"error,omitempty"`
}

// GetRecordDC contains a single Dublin Core record from GetRecord verb
type GetRecordDC struct {
	Record RecordDC `xml:"record"`
}

// DCMetadata represents extracted Dublin Core metadata
type DCMetadata struct {
	Title       []string `json:"title"`
	Creator     []string `json:"creator"`
	Subject     []string `json:"subject"`
	Description []string `json:"description"`
	Publisher   []string `json:"publisher"`
	Contributor []string `json:"contributor"`
	Date        []string `json:"date"`
	Type        []string `json:"type"`
	Format      []string `json:"format"`
	Identifier  []string `json:"identifier"`
	Source      []string `json:"source"`
	Language    []string `json:"language"`
	Relation    []string `json:"relation"`
	Coverage    []string `json:"coverage"`
	Rights      []string `json:"rights"`
}

// deduplicate removes duplicates from slice and returns unique values
func deduplicate(items []string) []string {
	if len(items) == 0 {
		return items
	}

	seen := make(map[string]bool)
	unique := []string{}

	for _, item := range items {
		if item == "" {
			continue
		}
		if !seen[item] {
			seen[item] = true
			unique = append(unique, item)
		}
	}

	return unique
}

// ExtractDCMetadata extracts Dublin Core metadata with deduplication
func (dc *DublinCore) ExtractDCMetadata() *DCMetadata {
	if dc == nil {
		return nil
	}

	return &DCMetadata{
		Title:       deduplicate(dc.Title),
		Creator:     deduplicate(dc.Creator),
		Subject:     deduplicate(dc.Subject),
		Description: deduplicate(dc.Description),
		Publisher:   deduplicate(dc.Publisher),
		Contributor: deduplicate(dc.Contributor),
		Date:        deduplicate(dc.Date),
		Type:        deduplicate(dc.Type),
		Format:      deduplicate(dc.Format),
		Identifier:  deduplicate(dc.Identifier),
		Source:      deduplicate(dc.Source),
		Language:    deduplicate(dc.Language),
		Relation:    deduplicate(dc.Relation),
		Coverage:    deduplicate(dc.Coverage),
		Rights:      deduplicate(dc.Rights),
	}
}

// ExtractAllDCMetadata extracts metadata from all Dublin Core records in OAI-PMH response
func (o *OAIPMHResponseDC) ExtractAllDCMetadata() []*DCMetadata {
	var results []*DCMetadata

	if o.ListRecords != nil {
		for _, record := range o.ListRecords.Records {
			if record.Metadata.DC != nil {
				metadata := record.Metadata.DC.ExtractDCMetadata()
				if metadata != nil {
					results = append(results, metadata)
				}
			}
		}
	}

	if o.GetRecord != nil {
		if o.GetRecord.Record.Metadata.DC != nil {
			metadata := o.GetRecord.Record.Metadata.DC.ExtractDCMetadata()
			if metadata != nil {
				results = append(results, metadata)
			}
		}
	}

	return results
}

// HarvestAllDC harvests all Dublin Core records using resumption tokens (backward compatible API)
func (c *OAIClient) HarvestAllDC(metadataPrefix string, callback func(*OAIPMHResponseDC) error) error {
	resumptionToken := ""

	for {
		resp, err := c.listRecordsRequestDC(metadataPrefix, resumptionToken)
		if err != nil {
			return err
		}

		// Type assert to concrete type for backward compatibility
		dcResp, ok := resp.(*OAIPMHResponseDC)
		if !ok {
			return fmt.Errorf("unexpected response type")
		}

		if err := callback(dcResp); err != nil {
			return fmt.Errorf("callback error: %w", err)
		}

		token := resp.GetResumptionToken()
		if token == "" {
			break
		}

		resumptionToken = token
	}

	return nil
}

// ParseOAIDCXML parses OAI-PMH XML data with Dublin Core metadata from bytes
func ParseOAIDCXML(data []byte) (*OAIPMHResponseDC, error) {
	var oaiResp OAIPMHResponseDC
	if err := xml.Unmarshal(data, &oaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	if oaiResp.Error != nil {
		return nil, fmt.Errorf("OAI-PMH error [%s]: %s", oaiResp.Error.Code, oaiResp.Error.Message)
	}

	return &oaiResp, nil
}

// Implement OAIResponse interface for OAIPMHResponseDC

// GetRecords returns all records in the response as MetadataExtractor interface
func (o *OAIPMHResponseDC) GetRecords() []MetadataExtractor {
	var extractors []MetadataExtractor

	if o.ListRecords != nil {
		for _, record := range o.ListRecords.Records {
			if record.Metadata.DC != nil {
				extractors = append(extractors, record.Metadata.DC)
			}
		}
	}

	if o.GetRecord != nil {
		if o.GetRecord.Record.Metadata.DC != nil {
			extractors = append(extractors, o.GetRecord.Record.Metadata.DC)
		}
	}

	return extractors
}

// GetResumptionToken returns the resumption token if available
func (o *OAIPMHResponseDC) GetResumptionToken() string {
	if o.ListRecords != nil && o.ListRecords.ResumptionToken != nil {
		return o.ListRecords.ResumptionToken.Token
	}
	return ""
}

// HasError returns true if the response contains an error
func (o *OAIPMHResponseDC) HasError() bool {
	return o.Error != nil
}

// GetError returns the error information
func (o *OAIPMHResponseDC) GetError() *OAIError {
	return o.Error
}

// Implement MetadataExtractor interface for DublinCore

// ExtractMetadata extracts metadata from Dublin Core record
func (dc *DublinCore) ExtractMetadata() interface{} {
	return dc.ExtractDCMetadata()
}

// GetFormat returns the metadata format type
func (dc *DublinCore) GetFormat() MetadataFormat {
	return FormatOAIDC
}
