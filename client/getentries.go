package client

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"

	ct "github.com/google/certificate-transparency/go"
)

// base64LeafEntry respresents a Base64 encoded leaf entry
type base64LeafEntry struct {
	LeafInput string `json:"leaf_input"`
	ExtraData string `json:"extra_data"`
}

// getEntriesReponse respresents the JSON response to the CT get-entries method
type getEntriesResponse struct {
	Entries []base64LeafEntry `json:"entries"` // the list of returned entries
}

// GetEntries attempts to retrieve the entries in the sequence [|start|, |end|] from the CT log server. (see section 4.6.)
// Returns a slice of LeafInputs or a non-nil error.
func (c *LogClient) GetEntries(start, end int64) ([]ct.LogEntry, error) {
	if end < 0 {
		return nil, errors.New("end should be >= 0")
	}
	if end < start {
		return nil, errors.New("start should be <= end")
	}
	var resp getEntriesResponse
	err := fetchAndParse(c.httpClient, fmt.Sprintf("%s%s?start=%d&end=%d", c.uri, GetEntriesPath, start, end), &resp)
	if err != nil {
		return nil, err
	}
	entries := make([]ct.LogEntry, len(resp.Entries))
	for index, entry := range resp.Entries {
		leafBytes, err := base64.StdEncoding.DecodeString(entry.LeafInput)
		leaf, err := ct.ReadMerkleTreeLeaf(bytes.NewBuffer(leafBytes))
		if err != nil {
			return nil, err
		}
		entries[index].Leaf = *leaf
		chainBytes, err := base64.StdEncoding.DecodeString(entry.ExtraData)

		var chain []ct.ASN1Cert
		switch leaf.TimestampedEntry.EntryType {
		case ct.X509LogEntryType:
			chain, err = ct.UnmarshalX509ChainArray(chainBytes)

		case ct.PrecertLogEntryType:
			chain, err = ct.UnmarshalPrecertChainArray(chainBytes)

		default:
			return nil, fmt.Errorf("saw unknown entry type: %v", leaf.TimestampedEntry.EntryType)
		}
		if err != nil {
			return nil, err
		}
		entries[index].Chain = chain
		entries[index].Index = start + int64(index)
	}
	return entries, nil
}