package jupiter

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed testdata/swapInstructionsWithJito.json
var swapInstructionsResponseJSON []byte

//go:embed testdata/swapInstructionsWithoutJito.json
var swapInstructionsResponseWithoutJitoJSON []byte

func TestSwapInstructionsResponse_Unmarshal(t *testing.T) {
	t.Run("parse swap instructions with jito tip", func(t *testing.T) {
		var response SwapInstructionsResponse

		err := json.Unmarshal(swapInstructionsResponseJSON, &response)
		require.NoError(t, err)

		require.Len(t, response.SetupInstructions, 4)
		require.NotEmpty(t, response.SwapInstruction)
		require.NotEmpty(t, response.CleanupInstruction)
		require.Len(t, response.OtherInstructions, 1)
	})

	t.Run("parse swap instructions without jito tip", func(t *testing.T) {
		var response SwapInstructionsResponse

		err := json.Unmarshal(swapInstructionsResponseWithoutJitoJSON, &response)
		require.NoError(t, err)

		require.Len(t, response.SetupInstructions, 4)
		require.NotEmpty(t, response.SwapInstruction)
		require.NotEmpty(t, response.CleanupInstruction)
		require.Empty(t, response.OtherInstructions)
	})
}
