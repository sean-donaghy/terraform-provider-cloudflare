package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/cloudflare/cloudflare-go"
)

func TestEmptyCacheDoesNotReturnRecord(t *testing.T) {

	ctx := context.Background()
	cache := NewDnsRecordCache()
	zoneID := "zone-A"
	recordID := "123"

	fetchAllRecordsInZone := func(zoneId string) ([]cloudflare.DNSRecord, error) {
		return []cloudflare.DNSRecord{}, nil
	}

	_, recordInCache := cache.DNSRecord(ctx, zoneID, recordID, fetchAllRecordsInZone)

	if recordInCache {
		t.Fatalf("Found record in empty cache")
	}
}

var zoneA = []cloudflare.DNSRecord{
	{
		ID:      "record-0",
		Type:    "A",
		Name:    "zero-A",
		Content: "127.0.0.0",
		ZoneID:  "zone-A",
	},
	{
		ID:      "record-1",
		Type:    "A",
		Name:    "one-A",
		Content: "127.0.0.1",
		ZoneID:  "zone-A",
	},
}

var zoneB = []cloudflare.DNSRecord{
	{
		ID:      "record-0",
		Type:    "A",
		Name:    "zero-B",
		Content: "127.0.0.0",
		ZoneID:  "zone-B",
	},
	{
		ID:      "record-1",
		Type:    "A",
		Name:    "one-B",
		Content: "127.0.0.1",
		ZoneID:  "zone-B",
	},
}

var fetchAllRecordsInZone = func(zoneId string) ([]cloudflare.DNSRecord, error) {
	switch zoneId {
	case "zone-A":
		return zoneA, nil
	case "zone-B":
		return zoneB, nil
	default:
		return []cloudflare.DNSRecord{}, error(fmt.Errorf("Unknown zone %s", zoneId))
	}
}

func TestFindRecordByRecordAndZoneIds(t *testing.T) {

	ctx := context.Background()
	cache := NewDnsRecordCache()

	findRecord := func(zoneId string, recordId string, expected cloudflare.DNSRecord) {

		record, recordInCache := cache.DNSRecord(ctx, zoneId, recordId, fetchAllRecordsInZone)

		t.Logf("Actual  : %s / %s | %s -> %s", record.ZoneID, record.ID, record.Name, record.Content)
		t.Logf("Expected: %s / %s | %s -> %s", expected.ZoneID, expected.ID, expected.Name, expected.Content)

		if recordInCache == false {
			t.Fatalf("Failed to find record: %s/%s", zoneId, recordId)
		}

		if record.ID != expected.ID {
			t.Fatalf("Found record with wrong Id")
		}

		if record.ZoneID != expected.ZoneID {
			t.Fatalf("Found record with wrong zoneId")
		}

		if record.Name != expected.Name {
			t.Fatalf("Found record with wrong Name")
		}

		if record.Content != expected.Content {
			t.Fatalf("Found record with wrong Content")
		}
	}

	findRecord("zone-A", "record-0", zoneA[0])
	findRecord("zone-A", "record-1", zoneA[1])
	findRecord("zone-B", "record-0", zoneB[0])
	findRecord("zone-B", "record-1", zoneB[1])
}

func TestFindsNoRecordForNonExistentZone(t *testing.T) {

	ctx := context.Background()
	cache := NewDnsRecordCache()

	_, recordInCache := cache.DNSRecord(ctx, "non-existent-zone", "record-0", fetchAllRecordsInZone)

	if recordInCache {
		t.Fatalf("Found record for non-existent zone")
	}
}
