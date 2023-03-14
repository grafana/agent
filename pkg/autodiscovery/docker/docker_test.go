package docker

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocker(t *testing.T) {
	res, err := Run()
	require.NoError(t, err)
	fmt.Println(res, err)
}
