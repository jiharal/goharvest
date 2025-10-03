package goharvest

import (
	"encoding/json"
	"fmt"
	"testing"
)

// testHarvestHelper is a helper function to reduce test code duplication
func testHarvestHelper(t *testing.T, oaiURL, sourceName string, recordIndex int, stopAfterFirst bool) {
	client := NewClient(oaiURL)
	totalRecords := 0
	var stopError error
	if stopAfterFirst {
		stopError = fmt.Errorf("stop harvesting for testing")
	}

	err := client.HarvestAllDC("oai_dc", func(resp *OAIPMHResponseDC) error {
		metadata := resp.ExtractAllDCMetadata()

		// Display sample record from first batch
		if totalRecords == 0 && len(metadata) > recordIndex {
			sample := metadata[recordIndex]
			jsonData, err := json.MarshalIndent(sample, "", "  ")
			if err != nil {
				return fmt.Errorf("error marshaling JSON: %w", err)
			}
			fmt.Printf("Sample DC metadata (%s):\n%s\n\n", sourceName, string(jsonData))
		}

		totalRecords += len(metadata)
		fmt.Printf("Processed %d records from %s...\n", totalRecords, sourceName)

		if stopError != nil {
			return stopError
		}
		return nil
	})

	// Check error
	if stopError != nil && err != nil {
		expectedErr := fmt.Sprintf("callback error: %s", stopError.Error())
		if err.Error() != expectedErr {
			t.Fatalf("Error harvesting: %v", err)
		}
	} else if err != nil {
		t.Fatalf("Error harvesting: %v", err)
	}

	fmt.Printf("\nTotal records harvested from %s: %d\n", sourceName, totalRecords)

	if totalRecords == 0 {
		t.Error("Expected at least one record")
	}
}

// TestOAIDCHarvestUAD demonstrates harvesting Dublin Core metadata from UAD EPrints
func TestOAIDCHarvestUAD(t *testing.T) {
	testHarvestHelper(t, "https://eprints.uad.ac.id/cgi/oai2", "UAD", 0, true)
}

// TestOAIDCHarvestUGM demonstrates harvesting Dublin Core metadata from UGM Lexicon journal
func TestOAIDCHarvestUGM(t *testing.T) {
	testHarvestHelper(t, "https://jurnal.ugm.ac.id/lexicon/oai", "UGM", 0, false)
}

// TestOAIDCHarvestUTDI demonstrates harvesting Dublin Core metadata from UTDI EPrints
func TestOAIDCHarvestUTDI(t *testing.T) {
	testHarvestHelper(t, "https://eprints.utdi.ac.id/cgi/oai2", "UTDI", 0, true)
}

// TestOAIDCHarvestUNY demonstrates harvesting Dublin Core metadata from UNY EPrints
func TestOAIDCHarvestUNY(t *testing.T) {
	testHarvestHelper(t, "https://eprints.uny.ac.id/cgi/oai2", "UNY", 1, true)
}

// TestOAIDCHarvestAMIKOM demonstrates harvesting Dublin Core metadata from AMIKOM journal
func TestOAIDCHarvestAMIKOM(t *testing.T) {
	testHarvestHelper(t, "https://jurnal.amikom.ac.id/index.php/joism/oai", "AMIKOM", 0, true)
}

// TestParseOAIDCXML demonstrates parsing Dublin Core XML from bytes
func TestParseOAIDCXML(t *testing.T) {
	client := NewClient("https://eprints.uad.ac.id/cgi/oai2")

	// Fetch one batch
	resp, err := client.listRecordsRequestDC("oai_dc", "")
	if err != nil {
		t.Fatalf("Error fetching data: %v", err)
	}

	// Cast to concrete type to extract metadata
	dcResp, ok := resp.(*OAIPMHResponseDC)
	if !ok {
		t.Fatal("Expected OAIPMHResponseDC")
	}

	allMetadata := dcResp.ExtractAllDCMetadata()

	if len(allMetadata) > 0 {
		jsonData, err := json.MarshalIndent(allMetadata[0], "", "  ")
		if err != nil {
			t.Fatalf("Error marshaling JSON: %v", err)
		}
		fmt.Printf("Extracted DC metadata:\n%s\n", string(jsonData))
	}

	fmt.Printf("\nTotal records extracted: %d\n", len(allMetadata))

	if len(allMetadata) == 0 {
		t.Error("Expected at least one record")
	}
}
