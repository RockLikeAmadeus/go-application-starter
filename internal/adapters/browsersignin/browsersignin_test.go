package browsersignin_test

import (
	"context"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/browsersignin"
)

func TestStart_ListensOnALoopbackPortAndHandsTheNativeSignInAddressToTheOpener(t *testing.T) {
	var opened string
	open := func(address string) error {
		opened = address
		return nil
	}

	attempt, err := browsersignin.Start(context.Background(), "https://app.example.com", open)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer attempt.Close()

	address, err := url.Parse(opened)
	if err != nil {
		t.Fatalf("opened address %q: %v", opened, err)
	}
	if address.Scheme != "https" || address.Host != "app.example.com" || address.Path != "/auth/native" {
		t.Errorf("opened %q, want https://app.example.com/auth/native", opened)
	}
	query := address.Query()
	if !regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`).MatchString(query.Get("state")) {
		t.Errorf("state %q is not 16 to 128 base64url characters", query.Get("state"))
	}
	if !regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`).MatchString(query.Get("challenge")) {
		t.Errorf("challenge %q is not 43 base64url characters", query.Get("challenge"))
	}
	port, err := strconv.Atoi(query.Get("port"))
	if err != nil || port < 1 || port > 65535 {
		t.Fatalf("port %q is not a port number", query.Get("port"))
	}
	conn, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1", query.Get("port")))
	if err != nil {
		t.Fatalf("nothing listens on 127.0.0.1:%d: %v", port, err)
	}
	conn.Close()
}
