package cluster

import (
	"reflect"
	"testing"
)

type MockedRand struct{}

// Shuffle reverse the array for a predictable outcome.
func (m *MockedRand) Shuffle(n int, swap func(i, j int)) {
	for i := 0; i < n/2; i++ {
		swap(i, n-i-1)
	}
}

func mockDiscoverPeers(peers []string, err error) func() ([]string, error) {
	return func() ([]string, error) {
		return peers, err
	}
}

func TestGetPeers(t *testing.T) {
	tests := []struct {
		name              string
		opts              Options
		expectedPeers     []string
		expectedError     error
		discoverPeersMock func() ([]string, error)
	}{
		{
			name:          "Test clustering disabled",
			opts:          Options{EnableClustering: false},
			expectedPeers: nil,
		},
		{
			name:          "Test no max peers limit",
			opts:          Options{EnableClustering: true, ClusterMaxInitJoinPeers: 0, DiscoverPeers: mockDiscoverPeers([]string{"A", "B"}, nil)},
			expectedPeers: []string{"A", "B"},
		},
		{
			name:          "Test max higher than number of peers",
			opts:          Options{EnableClustering: true, ClusterMaxInitJoinPeers: 5, DiscoverPeers: mockDiscoverPeers([]string{"A", "B", "C"}, nil)},
			expectedPeers: []string{"A", "B", "C"},
		},
		{
			name:          "Test max peers limit with shuffling",
			opts:          Options{EnableClustering: true, ClusterMaxInitJoinPeers: 2, DiscoverPeers: mockDiscoverPeers([]string{"A", "B", "C"}, nil)},
			expectedPeers: []string{"C", "B"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &Service{
				opts:    test.opts,
				randGen: &MockedRand{},
			}

			peers, _ := s.getPeers()

			if !reflect.DeepEqual(peers, test.expectedPeers) {
				t.Errorf("Expected peers %v, got %v", test.expectedPeers, peers)
			}
		})
	}
}
