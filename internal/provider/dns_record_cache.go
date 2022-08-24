package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type DnsRecordMap map[string]cloudflare.DNSRecord
type DnsZonesMap map[string]DnsRecordMap
type FetchAllRecordsInZone func(string) ([]cloudflare.DNSRecord, error)

type DnsRecordCache struct {
	mutex    sync.Mutex
	zonesMap DnsZonesMap
}

var dnsRecordCacheForAllZones = make(DnsZonesMap)
var dnsRecordCacheMutex sync.Mutex

func getDnsRecordFromCache(ctx context.Context, zoneID string, recordID string,
	fetchAllRecordsInZone FetchAllRecordsInZone, dnsRecordCacheForAllZones DnsZonesMap) (cloudflare.DNSRecord, bool) {

	dnsRecordCacheMutex.Lock()
	defer dnsRecordCacheMutex.Unlock()

	dnsRecordMap := ensureCacheForZoneIsInitialized(ctx, fetchAllRecordsInZone, zoneID)

	record, recordInCache := dnsRecordMap[recordID]

	return record, recordInCache
}

func ensureCacheForZoneIsInitialized(ctx context.Context, fetchAllRecordsInZone FetchAllRecordsInZone, zoneID string) DnsRecordMap {

	if dnsRecordCacheForZone, zoneFound := dnsRecordCacheForAllZones[zoneID]; zoneFound {
		tflog.Info(ctx, fmt.Sprintf("DNS zone %s already cached", zoneID))
		return dnsRecordCacheForZone
	}

	tflog.Warn(ctx, fmt.Sprintf("DNS zone %s not found in cache", zoneID))

	dnsRecordCacheForAllZones[zoneID] = make(DnsRecordMap)

	allDnsRecordsInZone, err := fetchAllRecordsInZone(zoneID)

	if err != nil {
		tflog.Warn(ctx, fmt.Sprintf("DNS zone %s - Failed to fetch records", zoneID))
		return dnsRecordCacheForAllZones[zoneID]
	}

	tflog.Info(ctx, fmt.Sprintf("DNS zone %s - Fetched all %d records", zoneID, len(allDnsRecordsInZone)))

	for i := 0; i < len(allDnsRecordsInZone); i++ {
		dnsRecord := allDnsRecordsInZone[i]
		dnsRecordCacheForAllZones[zoneID][dnsRecord.ID] = dnsRecord
	}

	tflog.Info(ctx, fmt.Sprintf("DNS zone %s - cache initialized", zoneID))

	return dnsRecordCacheForAllZones[zoneID]
}
