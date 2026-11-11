// Package instruction defines the Solana instruction shape returned by Jupiter APIs
// that build instructions instead of whole transactions.
package instruction

// Instruction is one Solana instruction with base64-encoded data.
type Instruction struct {
	ProgramID string        `json:"programId"`
	Accounts  []AccountMeta `json:"accounts"`
	Data      string        `json:"data"`
}

type AccountMeta struct {
	Pubkey     string `json:"pubkey"`
	IsSigner   bool   `json:"isSigner"`
	IsWritable bool   `json:"isWritable"`
}
