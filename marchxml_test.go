package goharvest

import (
	"encoding/xml"
	"os"
	"testing"
)

func TestHarvestAll(t *testing.T) {
	client := NewClient("https://balaiyanpus.jogjaprov.go.id/opac/oai")
	err := client.HarvestAll("marcxml", func(o *OAIPMHResponse) error {
		metadatas := o.ExtractAllBookMetadata()
		for _, x := range metadatas {
			t.Log(x)
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

func TestParseOAIPMHResponse(t *testing.T) {
	// Read the sample XML file
	data, err := os.ReadFile("testdata/sample_response.xml")
	if err != nil {
		t.Fatalf("Failed to read sample XML file: %v", err)
	}

	// Parse the XML
	var oaiResp OAIPMHResponse
	err = xml.Unmarshal(data, &oaiResp)
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}
	for _, x := range oaiResp.ListRecords.Records {
		t.Log(x.Metadata.MARCXML)
		t.Log("\n\n")
	}

	// Test top-level response
	t.Run("Response Structure", func(t *testing.T) {
		if oaiResp.ResponseDate == "" {
			t.Error("ResponseDate should not be empty")
		}
		if oaiResp.ResponseDate != "2025-10-02T10:05:19Z" {
			t.Errorf("Expected responseDate '2025-10-02T10:05:19Z', got '%s'", oaiResp.ResponseDate)
		}
	})

	// Test request information
	t.Run("Request Information", func(t *testing.T) {
		if oaiResp.Request.Verb != "ListRecords" {
			t.Errorf("Expected verb 'ListRecords', got '%s'", oaiResp.Request.Verb)
		}
		if oaiResp.Request.MetadataPrefix != "marcxml" {
			t.Errorf("Expected metadataPrefix 'marcxml', got '%s'", oaiResp.Request.MetadataPrefix)
		}
		if oaiResp.Request.URL != "http://balaiyanpus.jogjaprov.go.id/opac/index.php" {
			t.Errorf("Expected URL 'http://balaiyanpus.jogjaprov.go.id/opac/index.php', got '%s'", oaiResp.Request.URL)
		}
	})

	// Test ListRecords
	t.Run("ListRecords", func(t *testing.T) {
		if oaiResp.ListRecords == nil {
			t.Fatal("ListRecords should not be nil")
		}
		if len(oaiResp.ListRecords.Records) == 0 {
			t.Fatal("Records should not be empty")
		}
	})

	// Test first record
	t.Run("First Record", func(t *testing.T) {
		if len(oaiResp.ListRecords.Records) < 1 {
			t.Fatal("No records found")
		}

		record := oaiResp.ListRecords.Records[0]

		// Test header
		if record.Header.Identifier != "oai:balaiyanpus.jogjaprov.go.id:14" {
			t.Errorf("Expected identifier 'oai:balaiyanpus.jogjaprov.go.id:14', got '%s'", record.Header.Identifier)
		}

		// Test metadata and MARCXML
		if record.Metadata.MARCXML == nil {
			t.Fatal("MARCXML should not be nil")
		}
	})

	// Test MARC fields
	t.Run("MARC Fields", func(t *testing.T) {
		if len(oaiResp.ListRecords.Records) < 1 {
			t.Fatal("No records found")
		}

		marcRecord := oaiResp.ListRecords.Records[0].Metadata.MARCXML

		// Test control fields
		controlField001 := marcRecord.GetControlFieldValue("001")
		if controlField001 != "YOGYA000000000002408" {
			t.Errorf("Expected control field 001 value 'YOGYA000000000002408', got '%s'", controlField001)
		}

		controlField005 := marcRecord.GetControlFieldValue("005")
		if controlField005 != "20170404154010.0" {
			t.Errorf("Expected control field 005 value '20170404154010.0', got '%s'", controlField005)
		}

		// Test data fields
		title := marcRecord.GetFieldValue("245", "a")
		if title != "PANDUAN cerdas mahasiswa Jogja / editor, M. Solikhin, M. Farid" {
			t.Errorf("Expected title 'PANDUAN cerdas mahasiswa Jogja / editor, M. Solikhin, M. Farid', got '%s'", title)
		}

		publisher := marcRecord.GetFieldValue("260", "b")
		if publisher != "Kejora" {
			t.Errorf("Expected publisher 'Kejora', got '%s'", publisher)
		}

		year := marcRecord.GetFieldValue("260", "c")
		if year != "2005" {
			t.Errorf("Expected year '2005', got '%s'", year)
		}
	})

	// Test GetFieldValues (multiple values)
	t.Run("Multiple Field Values", func(t *testing.T) {
		if len(oaiResp.ListRecords.Records) < 1 {
			t.Fatal("No records found")
		}
		marcRecord := oaiResp.ListRecords.Records[0].Metadata.MARCXML
		// Get all 700 fields (authors)
		authors := marcRecord.GetFieldValues("700", "a")
		if len(authors) < 2 {
			t.Errorf("Expected at least 2 authors, got %d", len(authors))
		}
		if len(authors) > 0 && authors[0] != "M. Solikhin" {
			t.Errorf("Expected first author 'M. Solikhin', got '%s'", authors[0])
		}
		if len(authors) > 1 && authors[1] != "M. Farid" {
			t.Errorf("Expected second author 'M. Farid', got '%s'", authors[1])
		}

		// Get all 990 fields
		field990Values := marcRecord.GetFieldValues("990", "a")
		if len(field990Values) < 3 {
			t.Errorf("Expected at least 3 field 990 values, got %d", len(field990Values))
		}
	})

	// Test GetAllSubfields
	t.Run("Get All Subfields", func(t *testing.T) {
		if len(oaiResp.ListRecords.Records) < 1 {
			t.Fatal("No records found")
		}

		marcRecord := oaiResp.ListRecords.Records[0].Metadata.MARCXML

		// Get all 260 fields
		field260 := marcRecord.GetAllSubfields("260")
		if len(field260) == 0 {
			t.Error("Expected at least one 260 field")
		}

		if len(field260) > 0 {
			subfields := field260[0].Subfields
			if len(subfields) < 3 {
				t.Errorf("Expected at least 3 subfields in 260, got %d", len(subfields))
			}
		}
	})

	// Test indicators
	t.Run("Field Indicators", func(t *testing.T) {
		if len(oaiResp.ListRecords.Records) < 1 {
			t.Fatal("No records found")
		}

		marcRecord := oaiResp.ListRecords.Records[0].Metadata.MARCXML

		// Find 856 field to check indicators
		for _, field := range marcRecord.DataFields {
			if field.Tag == "856" {
				if field.Ind1 != "4" {
					t.Errorf("Expected ind1 '4' for tag 856, got '%s'", field.Ind1)
				}
				if field.Ind2 != "0" {
					t.Errorf("Expected ind2 '0' for tag 856, got '%s'", field.Ind2)
				}
				break
			}
		}
	})

	// Test multiple records
	t.Run("Multiple Records", func(t *testing.T) {
		if len(oaiResp.ListRecords.Records) < 2 {
			t.Skip("Skipping multiple records test, need at least 2 records")
		}

		// Test second record
		record2 := oaiResp.ListRecords.Records[1]
		if record2.Header.Identifier != "oai:balaiyanpus.jogjaprov.go.id:17" {
			t.Errorf("Expected second record identifier 'oai:balaiyanpus.jogjaprov.go.id:17', got '%s'", record2.Header.Identifier)
		}

		if record2.Metadata.MARCXML == nil {
			t.Error("Second record MARCXML should not be nil")
		}

		controlField001 := record2.Metadata.MARCXML.GetControlFieldValue("001")
		if controlField001 != "YOGYA-02090000041535" {
			t.Errorf("Expected second record control field 001 value 'YOGYA-02090000041535', got '%s'", controlField001)
		}
	})

	// Test no errors in response
	t.Run("No Errors", func(t *testing.T) {
		if oaiResp.Error != nil {
			t.Errorf("Expected no error in response, got error code: %s, message: %s", oaiResp.Error.Code, oaiResp.Error.Message)
		}
	})
}
