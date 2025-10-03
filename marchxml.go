package goharvest

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
)

// OAIClient represents an OAI-PMH client
type OAIClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new OAI-PMH client
func NewClient(baseURL string) *OAIClient {
	return &OAIClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// OAIPMHResponse represents the top-level OAI-PMH response
type OAIPMHResponse struct {
	XMLName         xml.Name         `xml:"OAI-PMH"`
	ResponseDate    string           `xml:"responseDate"`
	Request         OAIRequest       `xml:"request"`
	ListRecords     *ListRecords     `xml:"ListRecords,omitempty"`
	GetRecord       *GetRecord       `xml:"GetRecord,omitempty"`
	ListIdentifiers *ListIdentifiers `xml:"ListIdentifiers,omitempty"`
	Error           *OAIError        `xml:"error,omitempty"`
}

// OAIRequest represents the request information in the response
type OAIRequest struct {
	Verb            string `xml:"verb,attr"`
	MetadataPrefix  string `xml:"metadataPrefix,attr,omitempty"`
	ResumptionToken string `xml:"resumptionToken,attr,omitempty"`
	URL             string `xml:",chardata"`
}

// OAIError represents an OAI-PMH error
type OAIError struct {
	Code    string `xml:"code,attr"`
	Message string `xml:",chardata"`
}

// ListRecords contains the list of records from ListRecords verb
type ListRecords struct {
	Records         []Record         `xml:"record"`
	ResumptionToken *ResumptionToken `xml:"resumptionToken,omitempty"`
}

// GetRecord contains a single record from GetRecord verb
type GetRecord struct {
	Record Record `xml:"record"`
}

// ListIdentifiers contains the list of identifiers
type ListIdentifiers struct {
	Headers         []Header         `xml:"header"`
	ResumptionToken *ResumptionToken `xml:"resumptionToken,omitempty"`
}

// ResumptionToken for paginated results
type ResumptionToken struct {
	Token            string `xml:",chardata"`
	CompleteListSize int    `xml:"completeListSize,attr,omitempty"`
	Cursor           int    `xml:"cursor,attr,omitempty"`
	ExpirationDate   string `xml:"expirationDate,attr,omitempty"`
}

// Record represents an OAI-PMH record
type Record struct {
	Header   Header   `xml:"header"`
	Metadata Metadata `xml:"metadata"`
	About    *About   `xml:"about,omitempty"`
}

// Header contains record metadata
type Header struct {
	Status     string   `xml:"status,attr,omitempty"`
	Identifier string   `xml:"identifier"`
	DateStamp  string   `xml:"datestamp"`
	SetSpec    []string `xml:"setSpec,omitempty"`
}

// Metadata contains the actual record data
type Metadata struct {
	MARCXML *MARCRecord `xml:"record,omitempty"`
	Raw     []byte      `xml:",innerxml"`
}

// About contains optional about information
type About struct {
	Raw []byte `xml:",innerxml"`
}

// MARCRecord represents a MARCXML record
type MARCRecord struct {
	XMLName       xml.Name       `xml:"record"`
	Leader        string         `xml:"leader"`
	ControlFields []ControlField `xml:"controlfield"`
	DataFields    []DataField    `xml:"datafield"`
}

// ControlField represents a MARC control field (001-009)
type ControlField struct {
	Tag   string `xml:"tag,attr"`
	Value string `xml:",chardata"`
}

// DataField represents a MARC data field (010-999)
type DataField struct {
	Tag       string     `xml:"tag,attr"`
	Ind1      string     `xml:"ind1,attr"`
	Ind2      string     `xml:"ind2,attr"`
	Subfields []Subfield `xml:"subfield"`
}

// Subfield represents a MARC subfield
type Subfield struct {
	Code  string `xml:"code,attr"`
	Value string `xml:",chardata"`
}

// HarvestAll harvests all MARCXML records using resumption tokens (backward compatible API)
func (c *OAIClient) HarvestAll(metadataPrefix string, callback func(*OAIPMHResponse) error) error {
	resumptionToken := ""

	for {
		resp, err := c.listRecordsRequestMARCXML(metadataPrefix, resumptionToken, nil)
		if err != nil {
			return err
		}

		// Type assert to concrete type for backward compatibility
		marcResp, ok := resp.(*OAIPMHResponse)
		if !ok {
			return fmt.Errorf("unexpected response type")
		}

		if err := callback(marcResp); err != nil {
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

// ParseOAIPMHXML parses OAI-PMH XML data from bytes
func ParseOAIPMHXML(data []byte) (*OAIPMHResponse, error) {
	var oaiResp OAIPMHResponse
	if err := xml.Unmarshal(data, &oaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	if oaiResp.Error != nil {
		return nil, fmt.Errorf("OAI-PMH error [%s]: %s", oaiResp.Error.Code, oaiResp.Error.Message)
	}

	return &oaiResp, nil
}

// BookMetadata represents extracted bibliographic metadata from MARC record
type BookMetadata struct {
	RecordID        string   `json:"record_id"`        // 001
	LastModified    string   `json:"last_modified"`    // 005
	ISBN            string   `json:"isbn"`             // 020
	CallNumber      string   `json:"call_number"`      // 090
	MainAuthor      string   `json:"main_author"`      // 100
	CorporateAuthor string   `json:"corporate_author"` // 110
	MeetingName     string   `json:"meeting_name"`     // 111
	Title           string   `json:"title"`            // 245$a
	Subtitle        string   `json:"subtitle"`         // 245$b
	Responsibility  string   `json:"responsibility"`   // 245$c
	Edition         string   `json:"edition"`          // 250
	PublishPlace    string   `json:"publish_place"`    // 260$a
	Publisher       string   `json:"publisher"`        // 260$b
	PublishYear     string   `json:"publish_year"`     // 260$c
	PhysicalDesc    string   `json:"physical_desc"`    // 300
	Notes           []string `json:"notes"`            // 500
	Bibliography    string   `json:"bibliography"`     // 504
	Subjects        []string `json:"subjects"`         // 650
	Authors         []string `json:"authors"`          // 700
	Holdings        []string `json:"holdings"`         // 990, 999
	URL             string   `json:"url"`              // 856$u
	Classification  string   `json:"classification"`   // 082
}

// GetFieldValue retrieves the value of a specific MARC field and subfield
func (m *MARCRecord) GetFieldValue(tag, subfieldCode string) string {
	for _, field := range m.DataFields {
		if field.Tag == tag {
			for _, subfield := range field.Subfields {
				if subfield.Code == subfieldCode {
					return subfield.Value
				}
			}
		}
	}
	return ""
}

// GetFieldValues retrieves all values of a specific MARC field and subfield
func (m *MARCRecord) GetFieldValues(tag, subfieldCode string) []string {
	var values []string
	for _, field := range m.DataFields {
		if field.Tag == tag {
			for _, subfield := range field.Subfields {
				if subfield.Code == subfieldCode {
					values = append(values, subfield.Value)
				}
			}
		}
	}
	return values
}

// GetControlFieldValue retrieves the value of a control field
func (m *MARCRecord) GetControlFieldValue(tag string) string {
	for _, field := range m.ControlFields {
		if field.Tag == tag {
			return field.Value
		}
	}
	return ""
}

// GetAllSubfields retrieves all subfields for a given tag
func (m *MARCRecord) GetAllSubfields(tag string) []DataField {
	var fields []DataField
	for _, field := range m.DataFields {
		if field.Tag == tag {
			fields = append(fields, field)
		}
	}
	return fields
}

// ExtractBookMetadata extracts bibliographic metadata from a MARC record
func (m *MARCRecord) ExtractBookMetadata() *BookMetadata {
	if m == nil {
		return nil
	}

	metadata := &BookMetadata{
		Notes:    []string{},
		Subjects: []string{},
		Authors:  []string{},
		Holdings: []string{},
	}

	// Extract control fields
	metadata.RecordID = m.GetControlFieldValue("001")
	metadata.LastModified = m.GetControlFieldValue("005")

	// Extract ISBN (020)
	metadata.ISBN = m.GetFieldValue("020", "a")

	// Extract Classification (082)
	metadata.Classification = m.GetFieldValue("082", "a")

	// Extract Call Number (090)
	callNum090 := m.GetAllSubfields("090")
	if len(callNum090) > 0 {
		var callParts []string
		for _, subfield := range callNum090[0].Subfields {
			if subfield.Value != "" {
				callParts = append(callParts, subfield.Value)
			}
		}
		if len(callParts) > 0 {
			metadata.CallNumber = callParts[0]
			if len(callParts) > 1 {
				metadata.CallNumber += " " + callParts[1]
			}
		}
	}

	// Extract Main Author (100)
	metadata.MainAuthor = m.GetFieldValue("100", "a")

	// Extract Corporate Author (110)
	metadata.CorporateAuthor = m.GetFieldValue("110", "a")

	// Extract Meeting Name (111)
	metadata.MeetingName = m.GetFieldValue("111", "a")

	// Extract Title (245)
	metadata.Title = m.GetFieldValue("245", "a")
	metadata.Subtitle = m.GetFieldValue("245", "b")
	metadata.Responsibility = m.GetFieldValue("245", "c")

	// Extract Edition (250)
	metadata.Edition = m.GetFieldValue("250", "a")

	// Extract Publication info (260)
	metadata.PublishPlace = m.GetFieldValue("260", "a")
	metadata.Publisher = m.GetFieldValue("260", "b")
	metadata.PublishYear = m.GetFieldValue("260", "c")

	// Extract Physical Description (300)
	field300 := m.GetAllSubfields("300")
	if len(field300) > 0 {
		var physDesc []string
		for _, subfield := range field300[0].Subfields {
			if subfield.Value != "" {
				physDesc = append(physDesc, subfield.Value)
			}
		}
		if len(physDesc) > 0 {
			metadata.PhysicalDesc = physDesc[0]
			for i := 1; i < len(physDesc); i++ {
				metadata.PhysicalDesc += " " + physDesc[i]
			}
		}
	}

	// Extract Notes (500)
	metadata.Notes = m.GetFieldValues("500", "a")

	// Extract Bibliography (504)
	metadata.Bibliography = m.GetFieldValue("504", "a")

	// Extract Subjects (650)
	metadata.Subjects = m.GetFieldValues("650", "a")

	// Extract Additional Authors (700)
	metadata.Authors = m.GetFieldValues("700", "a")

	// Extract Holdings (990 and 999)
	holdings990 := m.GetFieldValues("990", "a")
	holdings999 := m.GetFieldValues("999", "a")
	metadata.Holdings = append(metadata.Holdings, holdings990...)
	metadata.Holdings = append(metadata.Holdings, holdings999...)

	// Extract URL (856)
	metadata.URL = m.GetFieldValue("856", "u")

	return metadata
}

// ExtractAllBookMetadata extracts metadata from all records in OAI-PMH response
func (o *OAIPMHResponse) ExtractAllBookMetadata() []*BookMetadata {
	var results []*BookMetadata

	if o.ListRecords != nil {
		for _, record := range o.ListRecords.Records {
			if record.Metadata.MARCXML != nil {
				metadata := record.Metadata.MARCXML.ExtractBookMetadata()
				if metadata != nil {
					results = append(results, metadata)
				}
			}
		}
	}

	if o.GetRecord != nil {
		if o.GetRecord.Record.Metadata.MARCXML != nil {
			metadata := o.GetRecord.Record.Metadata.MARCXML.ExtractBookMetadata()
			if metadata != nil {
				results = append(results, metadata)
			}
		}
	}

	return results
}

// Implement OAIResponse interface for OAIPMHResponse

// GetRecords returns all records in the response as MetadataExtractor interface
func (o *OAIPMHResponse) GetRecords() []MetadataExtractor {
	var extractors []MetadataExtractor

	if o.ListRecords != nil {
		for _, record := range o.ListRecords.Records {
			if record.Metadata.MARCXML != nil {
				extractors = append(extractors, record.Metadata.MARCXML)
			}
		}
	}

	if o.GetRecord != nil {
		if o.GetRecord.Record.Metadata.MARCXML != nil {
			extractors = append(extractors, o.GetRecord.Record.Metadata.MARCXML)
		}
	}

	return extractors
}

// GetResumptionToken returns the resumption token if available
func (o *OAIPMHResponse) GetResumptionToken() string {
	if o.ListRecords != nil && o.ListRecords.ResumptionToken != nil {
		return o.ListRecords.ResumptionToken.Token
	}
	return ""
}

// HasError returns true if the response contains an error
func (o *OAIPMHResponse) HasError() bool {
	return o.Error != nil
}

// GetError returns the error information
func (o *OAIPMHResponse) GetError() *OAIError {
	return o.Error
}

// Implement MetadataExtractor interface for MARCRecord

// ExtractMetadata extracts metadata from MARC record
func (m *MARCRecord) ExtractMetadata() interface{} {
	return m.ExtractBookMetadata()
}

// GetFormat returns the metadata format type
func (m *MARCRecord) GetFormat() MetadataFormat {
	return FormatMARCXML
}
