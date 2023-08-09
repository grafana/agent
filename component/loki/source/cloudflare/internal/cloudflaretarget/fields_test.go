package cloudflaretarget

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFields(t *testing.T) {
	tests := []struct {
		name         string
		fieldsType   FieldsType
		customFields []string
		expected     []string
	}{
		{
			name:         "Default fields",
			fieldsType:   FieldsTypeDefault,
			customFields: []string{},
			expected:     defaultFields,
		},
		{
			name:         "Custom fields",
			fieldsType:   FieldsTypeCustom,
			customFields: []string{"ClientIP", "OriginResponseBytes"},
			expected:     []string{"ClientIP", "OriginResponseBytes"},
		},
		{
			name:         "Default fields with added custom fields",
			fieldsType:   FieldsTypeDefault,
			customFields: []string{"WAFFlags", "WAFMatchedVar"},
			expected:     append(defaultFields, "WAFFlags", "WAFMatchedVar"),
		},
		{
			name:         "Default fields with duplicated custom fields",
			fieldsType:   FieldsTypeDefault,
			customFields: []string{"WAFFlags", "WAFFlags", "ClientIP"},
			expected:     append(defaultFields, "WAFFlags"), // clientIP is already part of defaultFields
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := Fields(test.fieldsType, test.customFields)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.expected, result)
		})
	}
}

func TestFindInvalidFields(t *testing.T) {
	invalidFields := []string{"InvalidField1", "InvalidField2"}

	result := FindInvalidFields(invalidFields)
	assert.ElementsMatch(t, invalidFields, result)
}
