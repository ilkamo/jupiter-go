package solanatx

import (
	"bytes"
	"encoding/base64"
	"testing"
)

// FuzzParse checks that arbitrary input never panics and that every accepted
// transaction re-encodes to exactly the input bytes and parses identically again.
func FuzzParse(f *testing.F) {
	f.Add(mustDecode(f, unsignedLegacy(f)))
	f.Add(mustDecode(f, unsignedVersion1(f)))
	f.Add([]byte{})
	f.Add([]byte{versionedMessagePrefix | version1Number})

	f.Fuzz(func(t *testing.T, wire []byte) {
		encoded := base64.StdEncoding.EncodeToString(wire)

		tx, err := Parse(encoded)
		if err != nil {
			return
		}

		if !bytes.Equal(tx.Bytes(), wire) || tx.Base64() != encoded {
			t.Fatal("accepted transaction does not round-trip")
		}

		again, err := Parse(tx.Base64())
		if err != nil {
			t.Fatalf("re-parse failed: %v", err)
		}

		if again.Version() != tx.Version() || len(again.RequiredSigners()) != len(tx.RequiredSigners()) {
			t.Fatal("re-parse changed the transaction")
		}
	})
}

func mustDecode(f *testing.F, encoded string) []byte {
	f.Helper()

	wire, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		f.Fatal(err)
	}

	return wire
}
