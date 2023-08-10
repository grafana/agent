package cloudflaretarget

// This code is copied from Promtail. The cloudflaretarget package is used to
// configure and run a target that can read from the Cloudflare Logpull API and
// forward entries to other loki components.

import (
	"fmt"
)

// FieldsType defines the set of fields to fetch alongside logs.
type FieldsType string

// Valid FieldsType values.
const (
	FieldsTypeDefault  FieldsType = "default"
	FieldsTypeMinimal  FieldsType = "minimal"
	FieldsTypeExtended FieldsType = "extended"
	FieldsTypeAll      FieldsType = "all"
	FieldsTypeCustom   FieldsType = "custom"
)

var (
	defaultFields = []string{
		"ClientIP", "ClientRequestHost", "ClientRequestMethod", "ClientRequestURI", "EdgeEndTimestamp", "EdgeResponseBytes",
		"EdgeRequestHost", "EdgeResponseStatus", "EdgeStartTimestamp", "RayID",
	}
	minimalFields = append(defaultFields, []string{
		"ZoneID", "ClientSSLProtocol", "ClientRequestProtocol", "ClientRequestPath", "ClientRequestUserAgent", "ClientRequestReferer",
		"EdgeColoCode", "ClientCountry", "CacheCacheStatus", "CacheResponseStatus", "EdgeResponseContentType", "SecurityLevel",
		"WAFAction", "WAFProfile", "WAFRuleID", "WAFRuleMessage", "EdgeRateLimitID", "EdgeRateLimitAction",
	}...)
	extendedFields = append(minimalFields, []string{
		"ClientSSLCipher", "ClientASN", "ClientIPClass", "CacheResponseBytes", "EdgePathingOp", "EdgePathingSrc", "EdgePathingStatus", "ParentRayID",
		"WorkerCPUTime", "WorkerStatus", "WorkerSubrequest", "WorkerSubrequestCount", "OriginIP", "OriginResponseStatus", "OriginSSLProtocol",
		"OriginResponseHTTPExpires", "OriginResponseHTTPLastModified",
	}...)
	allFields = append(extendedFields, []string{
		"BotScore", "BotScoreSrc", "ClientRequestBytes", "ClientSrcPort", "ClientXRequestedWith", "CacheTieredFill", "EdgeResponseCompressionRatio", "EdgeServerIP", "FirewallMatchesSources",
		"FirewallMatchesActions", "FirewallMatchesRuleIDs", "OriginResponseBytes", "OriginResponseTime", "ClientDeviceType", "WAFFlags", "WAFMatchedVar", "EdgeColoID",
		"RequestHeaders", "ResponseHeaders",
	}...)
)

// Fields returns the union of a set of fields represented by the Fieldtype and the given additional fields. The returned slice will contain no duplicates.
func Fields(t FieldsType, additionalFields []string) ([]string, error) {
	fieldsSubset, err := getFieldSubset(t)
	if err != nil {
		return nil, err
	}
	return mergeAndRemoveDuplicates(fieldsSubset, additionalFields), nil
}

func mergeAndRemoveDuplicates(fieldsSubset, additionalFields []string) []string {
	usedFields := make(map[string]struct{})
	var fields []string

	for _, field := range fieldsSubset {
		fields = append(fields, field)
		usedFields[field] = struct{}{}
	}

	for _, field := range additionalFields {
		if _, found := usedFields[field]; !found {
			fields = append(fields, field)
			usedFields[field] = struct{}{}
		}
	}
	return fields
}

// getFieldSubset returns the mapping of FieldsType to the set of fields it represents.
func getFieldSubset(t FieldsType) ([]string, error) {
	switch t {
	case FieldsTypeDefault:
		return defaultFields, nil
	case FieldsTypeMinimal:
		return minimalFields, nil
	case FieldsTypeExtended:
		return extendedFields, nil
	case FieldsTypeAll:
		return allFields, nil
	case FieldsTypeCustom:
		// Additional fields will be added later.
		return []string{}, nil
	default:
		return nil, fmt.Errorf("unknown fields type: %s", t)
	}
}
