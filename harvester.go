package goharvest

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

// Harvest is the unified entry point for harvesting OAI-PMH records
// It automatically detects the metadata format and returns appropriate parsers
func (c *OAIClient) Harvest(metadataPrefix string, callback HarvestCallback) error {
	format := MetadataFormat(metadataPrefix)

	switch format {
	case FormatMARCXML:
		return c.harvestMARCXML(metadataPrefix, callback)
	case FormatOAIDC:
		return c.harvestDublinCore(metadataPrefix, callback)
	default:
		return fmt.Errorf("unsupported metadata format: %s", metadataPrefix)
	}
}

// harvestMARCXML harvests MARCXML records
func (c *OAIClient) harvestMARCXML(metadataPrefix string, callback HarvestCallback) error {
	return c.harvestWithParser(metadataPrefix, c.listRecordsRequestMARCXML, callback)
}

// harvestDublinCore harvests Dublin Core records
func (c *OAIClient) harvestDublinCore(metadataPrefix string, callback HarvestCallback) error {
	return c.harvestWithParser(metadataPrefix, c.listRecordsRequestDC, callback)
}

// harvestWithParser is the unified harvest loop for all metadata formats
func (c *OAIClient) harvestWithParser(
	metadataPrefix string,
	parser func(string, string) (OAIResponse, error),
	callback HarvestCallback,
) error {
	resumptionToken := ""

	for {
		resp, err := parser(metadataPrefix, resumptionToken)
		if err != nil {
			return err
		}

		if err := callback(resp); err != nil {
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

// listRecordsRequestMARCXML performs a ListRecords request for MARCXML
func (c *OAIClient) listRecordsRequestMARCXML(metadataPrefix string, resumptionToken string) (OAIResponse, error) {
	body, err := c.performListRecordsRequest(metadataPrefix, resumptionToken)
	if err != nil {
		return nil, err
	}

	var oaiResp OAIPMHResponse
	if err := xml.Unmarshal(body, &oaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	if oaiResp.Error != nil {
		return nil, fmt.Errorf("OAI-PMH error [%s]: %s", oaiResp.Error.Code, oaiResp.Error.Message)
	}

	return &oaiResp, nil
}

// listRecordsRequestDC performs a ListRecords request for Dublin Core
func (c *OAIClient) listRecordsRequestDC(metadataPrefix string, resumptionToken string) (OAIResponse, error) {
	body, err := c.performListRecordsRequest(metadataPrefix, resumptionToken)
	if err != nil {
		return nil, err
	}

	var oaiResp OAIPMHResponseDC
	if err := xml.Unmarshal(body, &oaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	if oaiResp.Error != nil {
		return nil, fmt.Errorf("OAI-PMH error [%s]: %s", oaiResp.Error.Code, oaiResp.Error.Message)
	}

	return &oaiResp, nil
}

// performListRecordsRequest performs the actual HTTP request (unified logic)
func (c *OAIClient) performListRecordsRequest(metadataPrefix string, resumptionToken string) ([]byte, error) {
	url := c.BaseURL + "?verb=ListRecords"

	if resumptionToken != "" {
		url += "&resumptionToken=" + resumptionToken
	} else if metadataPrefix != "" {
		url += "&metadataPrefix=" + metadataPrefix
	} else {
		return nil, fmt.Errorf("either metadataPrefix or resumptionToken must be provided")
	}

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OAI data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}
