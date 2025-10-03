package goharvest

import (
	"testing"
)

// TestUnifiedHarvestMARCXML demonstrates using unified Harvest API with MARCXML
func TestUnifiedHarvestMARCXML(t *testing.T) {
	client := NewClient("https://balaiyanpus.jogjaprov.go.id/opac/oai")

	err := client.Harvest("marcxml", nil, func(response OAIResponse) error {
		// Get all records as MetadataExtractor interface
		records := response.GetRecords()

		t.Logf("Received %d records", len(records))

		for _, record := range records {
			// Extract metadata (returns interface{} that can be type asserted)
			metadata := record.ExtractMetadata()

			// Type assert to BookMetadata for MARCXML
			if bookMeta, ok := metadata.(*BookMetadata); ok {
				t.Logf("Title: %s", bookMeta.Title)
				t.Logf("Author: %s", bookMeta.MainAuthor)
			}
		}

		return nil
	})

	if err != nil {
		t.Errorf("Harvest failed: %v", err)
	}
}

// TestUnifiedHarvestDublinCore demonstrates using unified Harvest API with Dublin Core
func TestUnifiedHarvestDublinCore(t *testing.T) {
	// This is an example - replace with actual Dublin Core endpoint
	client := NewClient("https://example.com/oai")

	err := client.Harvest("oai_dc", nil, func(response OAIResponse) error {
		records := response.GetRecords()

		t.Logf("Received %d Dublin Core records", len(records))

		for _, record := range records {
			metadata := record.ExtractMetadata()

			// Type assert to DCMetadata for Dublin Core
			if dcMeta, ok := metadata.(*DCMetadata); ok {
				if len(dcMeta.Title) > 0 {
					t.Logf("DC Title: %s", dcMeta.Title[0])
				}
				if len(dcMeta.Creator) > 0 {
					t.Logf("DC Creator: %s", dcMeta.Creator[0])
				}
			}
		}

		return nil
	})

	if err != nil {
		// Skip if endpoint doesn't exist (this is just an example)
		t.Skipf("Skipped: %v", err)
	}
}

// TestUnifiedHarvestWithTypeSwitch demonstrates handling multiple formats dynamically
func TestUnifiedHarvestWithTypeSwitch(t *testing.T) {
	client := NewClient("https://balaiyanpus.jogjaprov.go.id/opac/oai")

	err := client.Harvest("marcxml", nil, func(response OAIResponse) error {
		records := response.GetRecords()

		for _, record := range records {
			// Get the format type
			format := record.GetFormat()
			t.Logf("Record format: %s", format)

			// Handle different formats
			switch format {
			case FormatMARCXML:
				metadata := record.ExtractMetadata()
				if bookMeta, ok := metadata.(*BookMetadata); ok {
					t.Logf("MARCXML - Title: %s", bookMeta.Title)
				}
			case FormatOAIDC:
				metadata := record.ExtractMetadata()
				if dcMeta, ok := metadata.(*DCMetadata); ok {
					if len(dcMeta.Title) > 0 {
						t.Logf("Dublin Core - Title: %s", dcMeta.Title[0])
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		t.Errorf("Harvest failed: %v", err)
	}
}

// TestBackwardCompatibilityMARCXML ensures old HarvestAll still works
func TestBackwardCompatibilityMARCXML(t *testing.T) {
	client := NewClient("https://balaiyanpus.jogjaprov.go.id/opac/oai")

	// Old API should still work
	err := client.HarvestAll("marcxml", func(o *OAIPMHResponse) error {
		metadatas := o.ExtractAllBookMetadata()
		t.Logf("Extracted %d book metadata records", len(metadatas))

		for _, meta := range metadatas {
			t.Logf("Title: %s", meta.Title)
		}

		return nil
	})

	if err != nil {
		t.Errorf("HarvestAll failed: %v", err)
	}
}

// TestBackwardCompatibilityDC ensures old HarvestAllDC still works
func TestBackwardCompatibilityDC(t *testing.T) {
	// This is an example - replace with actual Dublin Core endpoint
	client := NewClient("https://example.com/oai")

	// Old API should still work
	err := client.HarvestAllDC("oai_dc", func(o *OAIPMHResponseDC) error {
		metadatas := o.ExtractAllDCMetadata()
		t.Logf("Extracted %d DC metadata records", len(metadatas))

		return nil
	})

	if err != nil {
		t.Skipf("Skipped: %v", err)
	}
}

// TestHarvestWithDateRange demonstrates harvesting with date filtering
func TestHarvestWithDateRange(t *testing.T) {
	client := NewClient("https://balaiyanpus.jogjaprov.go.id/opac/oai")

	// Harvest records from January 2025
	dateRange := &DateRange{
		From:  "2025-01-01",
		Until: "2025-01-31",
	}

	recordCount := 0
	err := client.Harvest("marcxml", dateRange, func(response OAIResponse) error {
		records := response.GetRecords()
		recordCount += len(records)
		t.Logf("Received %d records in this batch", len(records))

		for _, record := range records {
			metadata := record.ExtractMetadata()
			if bookMeta, ok := metadata.(*BookMetadata); ok {
				t.Logf("Title: %s", bookMeta.Title)
			}
		}

		return nil
	})

	if err != nil {
		t.Errorf("Harvest with date range failed: %v", err)
	}

	t.Logf("Total records harvested in date range: %d", recordCount)
}

// TestHarvestWithFromDateOnly demonstrates harvesting from a specific date onwards
func TestHarvestWithFromDateOnly(t *testing.T) {
	client := NewClient("https://balaiyanpus.jogjaprov.go.id/opac/oai")

	// Harvest all records from October 2025 onwards
	dateRange := &DateRange{
		From: "2025-10-01",
	}

	err := client.Harvest("marcxml", dateRange, func(response OAIResponse) error {
		records := response.GetRecords()
		t.Logf("Received %d records from October 2025 onwards", len(records))
		return nil
	})

	if err != nil {
		t.Errorf("Harvest with from date failed: %v", err)
	}
}

// TestHarvestWithUntilDateOnly demonstrates harvesting up to a specific date
func TestHarvestWithUntilDateOnly(t *testing.T) {
	client := NewClient("https://balaiyanpus.jogjaprov.go.id/opac/oai")

	// Harvest all records up to December 2024
	dateRange := &DateRange{
		Until: "2024-12-31",
	}

	err := client.Harvest("oai_dc", dateRange, func(response OAIResponse) error {
		records := response.GetRecords()
		t.Logf("Received %d records up to December 2024", len(records))
		return nil
	})

	if err != nil {
		t.Skipf("Skipped: %v", err)
	}
}
