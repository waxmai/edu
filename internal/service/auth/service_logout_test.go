package auth

import (
	"context"
	"testing"
	"time"

	"edu-schedule-system/internal/proposal"
)

type stubTokenRevoker struct {
	revokedTokens []string
	revokedTTLs   []time.Duration
	revokedToken  string
	revokedTTL    time.Duration
	revokedMap    map[string]bool
	err           error
}

func (s *stubTokenRevoker) Revoke(ctx context.Context, token string, expiration time.Duration) error {
	if s.err != nil {
		return s.err
	}
	s.revokedToken = token
	s.revokedTokens = append(s.revokedTokens, token)
	s.revokedTTL = expiration
	s.revokedTTLs = append(s.revokedTTLs, expiration)
	if s.revokedMap == nil {
		s.revokedMap = map[string]bool{}
	}
	s.revokedMap[token] = true
	return nil
}

func (s *stubTokenRevoker) IsRevoked(ctx context.Context, token string) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	return s.revokedMap[token], nil
}

func TestLogoutRevokesToken(t *testing.T) {
	revoker := &stubTokenRevoker{}
	svc := New(nil, WithTokenRevoker(revoker))

	err := svc.Logout(context.Background(), testSession(), "token-123", "refresh-123")
	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if revoker.revokedToken != "refresh-123" {
		t.Fatalf("last revoked token = %q, want refresh-123", revoker.revokedToken)
	}
	if len(revoker.revokedTokens) != 2 || revoker.revokedTokens[0] != "token-123" || revoker.revokedTokens[1] != "refresh-123" {
		t.Fatalf("revoked tokens = %#v, want access then refresh", revoker.revokedTokens)
	}
	if len(revoker.revokedTTLs) != 2 || revoker.revokedTTLs[0] != logoutTokenTTL || revoker.revokedTTLs[1] != refreshTokenTTL {
		t.Fatalf("revoked ttls = %#v, want access/logout then refresh ttl", revoker.revokedTTLs)
	}
}

func TestLogoutRejectsEmptyToken(t *testing.T) {
	svc := New(nil)
	if err := svc.Logout(context.Background(), testSession(), "   ", ""); err == nil {
		t.Fatal("Logout() error = nil, want error")
	}
}

func testSession() proposal.SessionUserInfo {
	return proposal.SessionUserInfo{Id: 1, UserName: "admin", Status: proposal.UserStatusEnabled}
}
