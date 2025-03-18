package jupiter

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSwapInstructionsResponse_UnmarshalJSON(t *testing.T) {
	// Sample JSON response for SwapInstructionsResponse
	jsonData := `{
  "computeBudgetInstructions": [
    {
      "programId": "ComputeBudget111111111111111111111111111111",
      "accounts": [],
      "data": "AsBcFQA="
    },
    {
      "programId": "ComputeBudget111111111111111111111111111111",
      "accounts": [],
      "data": "AwcAAAAAAAAA"
    }
  ],
  "setupInstructions": [
    {
      "programId": "ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL",
      "accounts": [
        {
          "pubkey": "7R52mJ9sgGiwEGXBBm7ehQ5qz5uexpaJQtAkSUiwwQtH",
          "isSigner": true,
          "isWritable": true
        },
        {
          "pubkey": "DKf2QNuEq8Qb8QGtgujnw8WL1TLRtfPpMW5MnGn32PQa",
          "isSigner": false,
          "isWritable": true
        },
        {
          "pubkey": "7R52mJ9sgGiwEGXBBm7ehQ5qz5uexpaJQtAkSUiwwQtH",
          "isSigner": false,
          "isWritable": false
        },
        {
          "pubkey": "So11111111111111111111111111111111111111112",
          "isSigner": false,
          "isWritable": false
        },
        {
          "pubkey": "11111111111111111111111111111111",
          "isSigner": false,
          "isWritable": false
        },
        {
          "pubkey": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
          "isSigner": false,
          "isWritable": false
        }
      ],
      "data": "AQ=="
    },
    {
      "programId": "11111111111111111111111111111111",
      "accounts": [
        {
          "pubkey": "7R52mJ9sgGiwEGXBBm7ehQ5qz5uexpaJQtAkSUiwwQtH",
          "isSigner": true,
          "isWritable": true
        },
        {
          "pubkey": "DKf2QNuEq8Qb8QGtgujnw8WL1TLRtfPpMW5MnGn32PQa",
          "isSigner": false,
          "isWritable": true
        }
      ],
      "data": "AgAAAGA2HgAAAAAA"
    },
    {
      "programId": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
      "accounts": [
        {
          "pubkey": "DKf2QNuEq8Qb8QGtgujnw8WL1TLRtfPpMW5MnGn32PQa",
          "isSigner": false,
          "isWritable": true
        }
      ],
      "data": "EQ=="
    }
  ],
  "swapInstruction": {
    "programId": "JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4",
    "accounts": [
      {
        "pubkey": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
        "isSigner": false,
        "isWritable": false
      },
      {
        "pubkey": "7R52mJ9sgGiwEGXBBm7ehQ5qz5uexpaJQtAkSUiwwQtH",
        "isSigner": true,
        "isWritable": false
      },
      {
        "pubkey": "DKf2QNuEq8Qb8QGtgujnw8WL1TLRtfPpMW5MnGn32PQa",
        "isSigner": false,
        "isWritable": true
      },
      {
        "pubkey": "3Tx5eDWzMj2eGtRqzUBwH6wK3uwYn3QReM7goRDgqjpQ",
        "isSigner": false,
        "isWritable": true
      },
      {
        "pubkey": "JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4",
        "isSigner": false,
        "isWritable": false
      },
      {
        "pubkey": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
        "isSigner": false,
        "isWritable": false
      },
      {
        "pubkey": "JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4",
        "isSigner": false,
        "isWritable": false
      },
      {
        "pubkey": "D8cy77BBepLMngZx6ZukaTff5hCt1HrWyKk3Hnd9oitf",
        "isSigner": false,
        "isWritable": false
      },
      {
        "pubkey": "JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4",
        "isSigner": false,
        "isWritable": false
      },
      {
        "pubkey": "SoLFiHG9TfgtdUXUjWAxi3LtvYuFyDLVhBWxdMZxyCe",
        "isSigner": false,
        "isWritable": false
      },
      {
        "pubkey": "7R52mJ9sgGiwEGXBBm7ehQ5qz5uexpaJQtAkSUiwwQtH",
        "isSigner": false,
        "isWritable": false
      },
      {
        "pubkey": "DH4xmaWDnTzKXehVaPSNy9tMKJxnYL5Mo5U3oTHFtNYJ",
        "isSigner": false,
        "isWritable": true
      },
      {
        "pubkey": "5ep3LMR5gpCLD5KvSa9bnhR4R5Wm7HM7i1suP9u6ZvJT",
        "isSigner": false,
        "isWritable": true
      },
      {
        "pubkey": "3TokFuQgkkc6eLmafofNApdLkYpBvU1sZovyyScnQBD1",
        "isSigner": false,
        "isWritable": true
      },
      {
        "pubkey": "DKf2QNuEq8Qb8QGtgujnw8WL1TLRtfPpMW5MnGn32PQa",
        "isSigner": false,
        "isWritable": true
      },
      {
        "pubkey": "3Tx5eDWzMj2eGtRqzUBwH6wK3uwYn3QReM7goRDgqjpQ",
        "isSigner": false,
        "isWritable": true
      },
      {
        "pubkey": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
        "isSigner": false,
        "isWritable": false
      },
      {
        "pubkey": "Sysvar1nstructions1111111111111111111111111",
        "isSigner": false,
        "isWritable": false
      }
    ],
    "data": "5RfLl3rjrSoBAAAAPQBkAAFgNh4AAAAAAA7HAwAAAAAAMgAA"
  },
  "cleanupInstruction": {
    "programId": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
    "accounts": [
      {
        "pubkey": "DKf2QNuEq8Qb8QGtgujnw8WL1TLRtfPpMW5MnGn32PQa",
        "isSigner": false,
        "isWritable": true
      },
      {
        "pubkey": "7R52mJ9sgGiwEGXBBm7ehQ5qz5uexpaJQtAkSUiwwQtH",
        "isSigner": false,
        "isWritable": true
      },
      {
        "pubkey": "7R52mJ9sgGiwEGXBBm7ehQ5qz5uexpaJQtAkSUiwwQtH",
        "isSigner": true,
        "isWritable": false
      }
    ],
    "data": "CQ=="
  },
  "addressLookupTableAddresses": [
    "7KYzjjTydKxCSrjD3M3A2ntqKWtiGZszVX3ubA1FZcf5"
  ]
}`

	var response SwapInstructionsResponse
	err := json.Unmarshal([]byte(jsonData), &response)

	// Assert that unmarshaling was successful
	assert.NoError(t, err)

	// Assert that fields are properly populated
	assert.Equal(t, 1, len(response.AddressLookupTableAddresses))
	assert.Equal(t, "7KYzjjTydKxCSrjD3M3A2ntqKWtiGZszVX3ubA1FZcf5", response.AddressLookupTableAddresses[0])

	// // Check cleanup instruction
	// assert.NotNil(t, response.CleanupInstruction)
	// assert.Equal(t, "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA", response.CleanupInstruction.ProgramId)
	// assert.Equal(t, "CQ==", response.CleanupInstruction.Data)
	// assert.Equal(t, 1, len(response.CleanupInstruction.Accounts))
	// assert.Equal(t, "DKf2QNuEq8Qb8QGtgujnw8WL1TLRtfPpMW5MnGn32PQa", response.CleanupInstruction.Accounts[0].Pubkey)
	// assert.False(t, response.CleanupInstruction.Accounts[0].IsSigner)
	// assert.True(t, response.CleanupInstruction.Accounts[0].IsWritable)

	// // Check compute budget instructions
	// assert.Equal(t, 1, len(response.ComputeBudgetInstructions))
	// assert.Equal(t, "ComputeBudget111111111111111111111111111111", response.ComputeBudgetInstructions[0].ProgramId)
	// assert.Equal(t, "AsBcFQA=", response.ComputeBudgetInstructions[0].Data)
	// assert.Equal(t, 1, len(response.ComputeBudgetInstructions[0].Accounts))
	// assert.Equal(t, "ComputeBudget111111111111111111111111111111", response.ComputeBudgetInstructions[0].Accounts[0].Pubkey)
	// assert.False(t, response.ComputeBudgetInstructions[0].Accounts[0].IsSigner)
	// assert.True(t, response.ComputeBudgetInstructions[0].Accounts[0].IsWritable)

	// // Check setup instructions
	// assert.Equal(t, 1, len(response.SetupInstructions))
	// assert.Equal(t, "ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL", response.SetupInstructions[0].ProgramId)
	// assert.Equal(t, "AQ==", response.SetupInstructions[0].Data)
	// assert.Equal(t, 1, len(response.SetupInstructions[0].Accounts))
	// assert.Equal(t, "7R52mJ9sgGiwEGXBBm7ehQ5qz5uexpaJQtAkSUiwwQtH", response.SetupInstructions[0].Accounts[0].Pubkey)
	// assert.True(t, response.SetupInstructions[0].Accounts[0].IsSigner)
	// assert.False(t, response.SetupInstructions[0].Accounts[0].IsWritable)

	// // Check swap instruction
	// assert.Equal(t, "JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4", response.SwapInstruction.ProgramId)
	// assert.Equal(t, "5RfLl3rjrSoBAAAAPQBkAAFgNh4AAAAAAA7HAwAAAAAAMgAA", response.SwapInstruction.Data)
	// assert.Equal(t, 1, len(response.SwapInstruction.Accounts))
	// assert.Equal(t, "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA", response.SwapInstruction.Accounts[0].Pubkey)
	// assert.True(t, response.SwapInstruction.Accounts[0].IsSigner)
	// assert.True(t, response.SwapInstruction.Accounts[0].IsWritable)
}

func TestGetSwapInstructions(t *testing.T) {
	jupClient, err := NewClientWithResponses(DefaultAPIURL)
	// handle the error
	if err != nil {
		t.Fatalf("Failed to create Jupiter client: %v", err)
	}

	ctx := context.TODO()

	slippageBps := float32(250.0)

	// Get the current quote for a swap.
	// Ensure that the input and output mints are valid.
	// The amount is the smallest unit of the input token.
	quoteResponse, err := jupClient.GetQuoteWithResponse(ctx, &GetQuoteParams{
		InputMint:   "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
		OutputMint:  "So11111111111111111111111111111111111111112",
		Amount:      2000000,
		SlippageBps: &slippageBps,
	})
	// handle the error
	if err != nil {
		t.Fatalf("Failed to get quote: %v", err)
	}

	quote := quoteResponse.JSON200

	// Define the prioritization fee in lamports.
	prioritizationFeeLamports := &struct {
		JitoTipLamports              *int `json:"jitoTipLamports,omitempty"`
		PriorityLevelWithMaxLamports *struct {
			MaxLamports   *int    `json:"maxLamports,omitempty"`
			PriorityLevel *string `json:"priorityLevel,omitempty"`
		} `json:"priorityLevelWithMaxLamports,omitempty"`
	}{
		PriorityLevelWithMaxLamports: &struct {
			MaxLamports   *int    `json:"maxLamports,omitempty"`
			PriorityLevel *string `json:"priorityLevel,omitempty"`
		}{
			MaxLamports:   new(int),
			PriorityLevel: new(string),
		},
	}

	*prioritizationFeeLamports.PriorityLevelWithMaxLamports.MaxLamports = 1000
	*prioritizationFeeLamports.PriorityLevelWithMaxLamports.PriorityLevel = "high"

	// If you prefer to set a Jito tip, you can use the following line instead of the above block.
	// *prioritizationFeeLamports.JitoTipLamports = 1000

	dynamicComputeUnitLimit := true
	// Get instructions for a swap.
	// Ensure your public key is valid.
	swapResponse, err := jupClient.PostSwapInstructionsWithResponse(ctx, PostSwapInstructionsJSONRequestBody{
		PrioritizationFeeLamports: prioritizationFeeLamports,
		QuoteResponse:             *quote,
		UserPublicKey:             "7R52mJ9sgGiwEGXBBm7ehQ5qz5uexpaJQtAkSUiwwQtH",
		DynamicComputeUnitLimit:   &dynamicComputeUnitLimit,
	})
	// handle the error
	if err != nil {
		t.Fatalf("Failed to get swap instructions: %v", err)
	}
	swap := swapResponse.JSON200
	fmt.Println(swap)
	t.Log(swap)
}
