package elf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGoBuildID(t *testing.T) {
	testcases := []struct {
		f   string
		typ string
		id  string
	}{
		{"go12", "", ""},
		{"go12-static", "", ""},
		{"go16", "go", "92fNM3QG6qxprToyNKC6/aWDMs1NbNgS6utRcIALh/OnWrTUXnoISctQzMsAsP/X49oo_6m-sncjFUHr8D7"},
		{"go16-static", "go", "87M3bytdaVMnfLcC6UiX/aWDMs1NbNgS6utRcIALh/OnWrTUXnoISctQzMsAsP/X49oo_6m-sncjFUHr8D7"},
		{"go18", "go", "OQZSseDjmXatQp02XfSs/YnxZZfCwiY5UEDTfATTP/ZvesUZonVPjE1Be6F2gG/ET_kcbC7pPCzH8jfF77U"},
		{"go18-static", "go", "bAg3h9kqLtwBN5WyJ-8P/e2Jj53oDvybTWnT59xY-/ZvesUZonVPjE1Be6F2gG/pDIuRQ8JQXQfEmRf_0-O"},
		{"go20", "go", "Qffu__H4fOgskLcG9xgZ/IJ9VzAiPxBstzejlZLmK/EGJHHzTL5Vs7GGklz10L/r6shgckObuxZ4kbw9YGX"},
		{"go20-static", "go", "otBDhR4Gpy9N5G-o7ltP/UAAPpyRPxBegdj7J8l7o/EGJHHzTL5Vs7GGklz10L/oyXFW3eY5aQbuMp6Y-up"},
		{"elf", "gnu", "1fcfa068c5fdb9f31e6d9f3f89019beacb70182d"},
		{"elf.debug", "gnu", "1fcfa068c5fdb9f31e6d9f3f89019beacb70182d"},
		{"elf.stripped", "gnu", "1fcfa068c5fdb9f31e6d9f3f89019beacb70182d"},
	}
	for _, testcase := range testcases {
		t.Run(testcase.f, func(t *testing.T) {
			me, err := NewMMapedElfFile("./testdata/elfs/" + testcase.f)
			require.NoError(t, err)
			defer me.Close()
			id, err := me.BuildID()
			if testcase.id == "" {
				require.Error(t, err)
				require.Empty(t, id)
			} else {
				require.NoError(t, err)
				require.Equal(t, testcase.id, id.ID())
				require.Equal(t, testcase.typ, id.Type())
			}
		})
	}
}
