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

type DnsRecordCacheStats struct {
	requests  int
	cacheHits int
}

type DnsRecordCache struct {
	mutex    sync.Mutex
	zonesMap DnsZonesMap
	stats    DnsRecordCacheStats
}

func NewDnsRecordCache() *DnsRecordCache {
	c := new(DnsRecordCache)
	c.zonesMap = make(DnsZonesMap)
	return c
}

func (cache *DnsRecordCache) DNSRecord(ctx context.Context, zoneID string, recordID string,
	fetchAllRecordsInZone FetchAllRecordsInZone) (cloudflare.DNSRecord, bool) {

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	dnsRecordMap := cache.ensureCacheForZoneIsInitialized(ctx, fetchAllRecordsInZone, zoneID)

	record, recordInCache := dnsRecordMap[recordID]

	tflog.Trace(ctx, fmt.Sprintf("DNS zone %s / record %s in cache ? %t", zoneID, recordID, recordInCache))

	cache.stats.requests += 1
	if recordInCache {
		cache.stats.cacheHits += 1
	}

	tflog.Trace(ctx, fmt.Sprintf("DNS cache stats: %4d cacheHits / %4d requests", cache.stats.cacheHits, cache.stats.requests))

	return record, recordInCache
}

func (cache *DnsRecordCache) ensureCacheForZoneIsInitialized(ctx context.Context, fetchAllRecordsInZone FetchAllRecordsInZone, zoneID string) DnsRecordMap {

	if dnsRecordCacheForZone, zoneFound := cache.zonesMap[zoneID]; zoneFound {
		tflog.Info(ctx, fmt.Sprintf("DNS zone %s already cached", zoneID))
		return dnsRecordCacheForZone
	}

	tflog.Warn(ctx, fmt.Sprintf("DNS zone %s not found in cache", zoneID))

	cache.zonesMap[zoneID] = make(DnsRecordMap)

	allDnsRecordsInZone, err := fetchAllRecordsInZone(zoneID)

	if err != nil {
		tflog.Warn(ctx, fmt.Sprintf("DNS zone %s - Failed to fetch records", zoneID))
		return cache.zonesMap[zoneID]
	}

	tflog.Info(ctx, fmt.Sprintf("DNS zone %s - Fetched all %d records", zoneID, len(allDnsRecordsInZone)))

	for i := 0; i < len(allDnsRecordsInZone); i++ {
		dnsRecord := allDnsRecordsInZone[i]
		cache.zonesMap[zoneID][dnsRecord.ID] = dnsRecord
	}

	tflog.Info(ctx, fmt.Sprintf("DNS zone %s - cache initialized", zoneID))

	return cache.zonesMap[zoneID]
}
