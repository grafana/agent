package armclient

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/webdevops/go-common/utils/to"
)

const (
	AzurePrometheusLabelPrefix = "tag_"
)

var (
	azureTagNameToPrometheusNameRegExp = regexp.MustCompile("[^_a-zA-Z0-9]")
)

func AddResourceTagsToPrometheusLabelsDefinition(labels, tags []string) []string {
	return AddResourceTagsToPrometheusLabelsDefinitionWithCustomPrefix(labels, tags, AzurePrometheusLabelPrefix)
}

func AddResourceTagsToPrometheusLabelsDefinitionWithCustomPrefix(labels, tags []string, labelPrefix string) []string {
	for _, val := range tags {
		tagName := labelPrefix + val
		labels = append(labels, tagName)
	}

	return labels
}

func AddResourceTagsToPrometheusLabels(labels prometheus.Labels, resourceTags interface{}, tags []string) prometheus.Labels {
	return AddResourceTagsToPrometheusLabelsWithCustomPrefix(labels, resourceTags, tags, AzurePrometheusLabelPrefix)
}

func AddResourceTagsToPrometheusLabelsWithCustomPrefix(labels prometheus.Labels, resourceTags interface{}, tags []string, labelPrefix string) prometheus.Labels {
	resourceTagMap := translateTagsToStringMap(resourceTags)

	// normalize
	resourceTagMap = normalizeTags(resourceTagMap)

	for _, tag := range tags {
		tagParts := strings.SplitN(tag, "?", 2)
		tag = tagParts[0]

		tagSettings := ""
		if len(tagParts) == 2 {
			tagSettings = tagParts[1]
		}

		tag = strings.ToLower(tag)
		tagLabel := labelPrefix + azureTagNameToPrometheusNameRegExp.ReplaceAllLiteralString(tag, "_")
		labels[tagLabel] = ""

		if val, exists := resourceTagMap[tag]; exists {
			if tagSettings != "" {
				val = applyTagValueSettings(val, tagSettings)
			}
			labels[tagLabel] = val
		}
	}

	return labels
}

func applyTagValueSettings(val string, settings string) string {
	ret := val
	settingQuery, _ := url.ParseQuery(settings)

	if settingQuery.Has("toLower") || settingQuery.Has("tolower") {
		ret = strings.ToLower(ret)
	}

	if settingQuery.Has("toUpper") || settingQuery.Has("toupper") {
		ret = strings.ToUpper(ret)
	}

	return ret
}

func normalizeTags(tags map[string]string) map[string]string {
	ret := map[string]string{}

	for tagName, tagValue := range tags {
		tagName = strings.ToLower(tagName)
		ret[tagName] = strings.TrimSpace(tagValue)
	}

	return ret
}

func translateTagsToStringMap(tags interface{}) map[string]string {
	switch v := tags.(type) {
	case map[string]*string:
		return to.StringMap(v)
	case map[string]string:
		return v
	}

	return map[string]string{}
}
