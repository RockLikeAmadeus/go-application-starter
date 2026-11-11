package system_test

import (
	"encoding/base64"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/system"
)

func TestTokensMakesA32ByteRandomTokenEachTime(t *testing.T) {
	var tokens system.Tokens

	first, second := tokens.NewToken(), tokens.NewToken()

	raw, err := base64.RawURLEncoding.DecodeString(first)
	if err != nil {
		t.Fatalf("token %q is not unpadded base64url: %v", first, err)
	}
	if len(raw) != 32 {
		t.Errorf("token holds %d bytes, want 32", len(raw))
	}
	if first == second {
		t.Errorf("two tokens are both %q, want them different", first)
	}
}
