package metrics

import (
	"fmt"
	"net/url"
	"strings"
)

func paramsGetWithDefault(params url.Values, name, defaultValue string) (value string) {
	value = params.Get(name)
	if value == "" {
		value = defaultValue
	}
	return
}

func paramsGetList(params url.Values, name string) (list []string, err error) {
	for _, v := range params[name] {
		list = append(list, stringToStringList(v, ",")...)
	}
	return
}

func paramsGetListRequired(params url.Values, name string) (list []string, err error) {
	list, err = paramsGetList(params, name)

	if len(list) == 0 {
		err = fmt.Errorf("parameter \"%v\" is missing", name)
		return
	}

	return
}

func stringToStringList(v string, sep string) (list []string) {
	for _, v := range strings.Split(v, sep) {
		list = append(list, strings.TrimSpace(v))
	}
	return
}
