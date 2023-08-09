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
	allFieldsMap = buildAllFieldsMap(allFields)
)

// Fields returns the mapping of FieldsType to the set of fields it represents.
func Fields(t FieldsType) ([]string, error) {
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
		return []string{}, nil
	default:
		return nil, fmt.Errorf("unknown fields type: %s", t)
	}
}

func buildAllFieldsMap(allFields []string) map[string]bool {
	fieldsMap := make(map[string]bool)
	for _, field := range allFields {
		fieldsMap[field] = true
	}
	return fieldsMap
}

func FindInvalidFields(fields []string) []string {
	var invalidFields []string

	for _, field := range fields {
		if !allFieldsMap[field] {
			invalidFields = append(invalidFields, field)
		}
	}
	return invalidFields
}
