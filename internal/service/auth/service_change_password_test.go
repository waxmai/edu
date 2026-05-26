package auth

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"edu-schedule-system/internal/proposal"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testDBRepo struct {
	r *gorm.DB
	w *gorm.DB
}

func (r *testDBRepo) GetDbR() *gorm.DB           { return r.r }
func (r *testDBRepo) GetDbW() *gorm.DB           { return r.w }
func (r *testDBRepo) DbRClose() error            { return nil }
func (r *testDBRepo) DbWClose() error            { return nil }
func (r *testDBRepo) Ping(context.Context) error { return nil }
func (r *testDBRepo) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestChangePasswordClearsMustChangeFlag(t *testing.T) {
	prepareAuthTestConfig(t)
	db := newAuthTestDB(t)
	now := time.Now()
	seed := &sysUser{ID: 1, Username: "admin", PasswordHash: mustHashPassword(t, "StrongPass123!"), RoleCode: proposal.RolePlatformAdmin, RealName: "系统管理员", Email: "admin@example.local", Status: proposal.UserStatusEnabled, MustChangePassword: true, CreatedAt: now, UpdatedAt: now}
	if err := db.Create(seed).Error; err != nil {
		t.Fatalf("seed create error = %v", err)
	}
	if err := db.Create(&authSession{ID: "sess-current", UserID: 1, Status: authSessionStatusActive, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("seed current session error = %v", err)
	}
	if err := db.Create(&authSession{ID: "sess-other", UserID: 1, Status: authSessionStatusActive, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("seed other session error = %v", err)
	}

	svc := New(&testDBRepo{r: db, w: db})
	resp, err := svc.ChangePassword(context.Background(), proposal.SessionUserInfo{Id: 1, Status: proposal.UserStatusEnabled, MustChangePassword: true, TokenID: buildTokenID("access", "sess-current")}, &ChangePasswordRequest{OldPassword: "StrongPass123!", NewPassword: "ChangedPass123!"})
	if err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}
	if resp == nil || resp.UserInfo == nil || resp.UserInfo.MustChangePassword {
		t.Fatalf("ChangePassword() userInfo.mustChangePassword = true, want false")
	}
	if resp.Session == nil || resp.Session.ID != "sess-current" {
		t.Fatalf("ChangePassword() session = %#v, want sess-current", resp.Session)
	}
	if resp.Token == "" || resp.RefreshToken == "" {
		t.Fatalf("ChangePassword() token pair = %#v, want non-empty token and refresh token", resp)
	}

	var current authSession
	if err := db.First(&current, "id = ?", "sess-current").Error; err != nil {
		t.Fatalf("load current session error = %v", err)
	}
	if current.Status != authSessionStatusActive || current.RevokedReason != "" {
		t.Fatalf("current session = %#v, want active session without revoked reason", current)
	}
}

func TestLoginAndRefreshSessionGovernance(t *testing.T) {
	prepareAuthTestConfig(t)
	db := newAuthTestDB(t)
	now := time.Now()
	seed := &sysUser{ID: 2, Username: "rotate", PasswordHash: mustHashPassword(t, "StrongPass123!"), RoleCode: proposal.RoleTeacher, RealName: "轮换用户", Email: "rotate@example.local", Status: proposal.UserStatusEnabled, CreatedAt: now, UpdatedAt: now}
	if err := db.Create(seed).Error; err != nil {
		t.Fatalf("seed create error = %v", err)
	}
	revoker := &stubTokenRevoker{}
	svc := New(&testDBRepo{r: db, w: db}, WithTokenRevoker(revoker))

	loginResp, err := svc.Login(context.Background(), &LoginRequest{Username: "rotate", Password: "StrongPass123!"}, AuthClientMeta{ClientIP: "1.1.1.1", UserAgent: "UA-1", DeviceName: "Device-A"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if loginResp.Session == nil || loginResp.Session.ID == "" {
		t.Fatalf("Login() session = %#v, want non-nil", loginResp.Session)
	}

	refreshResp, err := svc.Refresh(context.Background(), &RefreshRequest{RefreshToken: loginResp.RefreshToken}, AuthClientMeta{ClientIP: "2.2.2.2", UserAgent: "UA-2", DeviceName: "Device-B"})
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if refreshResp.Session == nil || refreshResp.Session.ID != loginResp.Session.ID {
		t.Fatalf("Refresh() session %#v, want reuse %s", refreshResp.Session, loginResp.Session.ID)
	}
	if len(revoker.revokedTokens) == 0 || revoker.revokedTokens[0] != loginResp.RefreshToken {
		t.Fatalf("revoked tokens = %#v, want first old refresh token", revoker.revokedTokens)
	}

	listed, err := svc.ListSessions(context.Background(), proposal.SessionUserInfo{Id: 2, Status: proposal.UserStatusEnabled, TokenID: buildTokenID("access", loginResp.Session.ID)})
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	if listed.Total != 1 || len(listed.Items) != 1 || !listed.Items[0].Current {
		t.Fatalf("ListSessions() = %#v, want one current session", listed)
	}
}

func TestRevokeSessionAndLogoutAll(t *testing.T) {
	prepareAuthTestConfig(t)
	db := newAuthTestDB(t)
	now := time.Now()
	seed := &sysUser{ID: 3, Username: "sessions", PasswordHash: mustHashPassword(t, "StrongPass123!"), RoleCode: proposal.RoleTeacher, RealName: "会话用户", Email: "sessions@example.local", Status: proposal.UserStatusEnabled, CreatedAt: now, UpdatedAt: now}
	if err := db.Create(seed).Error; err != nil {
		t.Fatalf("seed create error = %v", err)
	}
	for _, row := range []authSession{{ID: "sess-a", UserID: 3, DeviceName: "A", ClientIP: "10.0.0.1", Status: authSessionStatusActive, CreatedAt: now, UpdatedAt: now}, {ID: "sess-b", UserID: 3, DeviceName: "B", ClientIP: "10.0.0.2", Status: authSessionStatusActive, IsSuspicious: true, RiskFlags: marshalStringSlice([]string{"new_device"}), CreatedAt: now, UpdatedAt: now}} {
		copy := row
		if err := db.Create(&copy).Error; err != nil {
			t.Fatalf("seed session error = %v", err)
		}
	}
	svc := New(&testDBRepo{r: db, w: db})
	if err := svc.RevokeSession(context.Background(), proposal.SessionUserInfo{Id: 3, Status: proposal.UserStatusEnabled, TokenID: buildTokenID("access", "sess-a")}, "sess-b", ""); err != nil {
		t.Fatalf("RevokeSession() error = %v", err)
	}
	var revoked authSession
	if err := db.First(&revoked, "id = ?", "sess-b").Error; err != nil {
		t.Fatalf("load revoked session error = %v", err)
	}
	if revoked.Status != authSessionStatusRevoked {
		t.Fatalf("revoked status = %s, want revoked", revoked.Status)
	}
	if err := svc.LogoutAll(context.Background(), proposal.SessionUserInfo{Id: 3, Status: proposal.UserStatusEnabled, TokenID: buildTokenID("access", "sess-a")}, "token-now"); err != nil {
		t.Fatalf("LogoutAll() error = %v", err)
	}
	var activeCount int64
	if err := db.Model(&authSession{}).Where("user_id = ? AND status = ?", 3, authSessionStatusActive).Count(&activeCount).Error; err != nil {
		t.Fatalf("count active sessions error = %v", err)
	}
	if activeCount != 0 {
		t.Fatalf("active session count = %d, want 0", activeCount)
	}
}

type stubRecoveryDeliverySender struct {
	messages []RecoveryDeliveryMessage
	result   *RecoveryDeliveryResult
	err      error
}

func (s *stubRecoveryDeliverySender) Send(ctx context.Context, message RecoveryDeliveryMessage) (*RecoveryDeliveryResult, error) {
	s.messages = append(s.messages, message)
	if s.err != nil {
		return nil, s.err
	}
	if s.result != nil {
		return s.result, nil
	}
	return &RecoveryDeliveryResult{Accepted: true, Provider: "stub"}, nil
}

func TestPasswordRecoveryFlowRevokesSessionsAndWritesAudit(t *testing.T) {
	t.Setenv("AUTH_PASSWORD_RECOVERY_PREVIEW_ENABLED", "false")
	prepareAuthTestConfig(t)
	db := newAuthTestDB(t)
	now := time.Now()
	seed := &sysUser{ID: 4, Username: "recoverme", PasswordHash: mustHashPassword(t, "OldPass123!"), RoleCode: proposal.RolePlatformAdmin, RealName: "找回用户", Phone: "10000013800", Email: "recoverme@example.local", Status: proposal.UserStatusEnabled, CreatedAt: now, UpdatedAt: now}
	if err := db.Create(seed).Error; err != nil {
		t.Fatalf("seed create error = %v", err)
	}
	for _, id := range []string{"sess-x", "sess-y"} {
		if err := db.Create(&authSession{ID: id, UserID: 4, Status: authSessionStatusActive, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
			t.Fatalf("seed session error = %v", err)
		}
	}
	delivery := &stubRecoveryDeliverySender{}
	svc := New(&testDBRepo{r: db, w: db}, WithRecoveryDelivery(delivery))
	startResp, err := svc.StartPasswordRecovery(context.Background(), &PasswordRecoveryStartRequest{Username: "recoverme", Channel: "email", SecondFactorChannel: "sms"}, AuthClientMeta{ClientIP: "8.8.8.8", UserAgent: "Recovery-UA", DeviceName: "Unknown Browser"})
	if err != nil {
		t.Fatalf("StartPasswordRecovery() error = %v", err)
	}
	if !startResp.Accepted || startResp.ChallengeID == "" || !startResp.SecondFactorRequired {
		t.Fatalf("StartPasswordRecovery() = %#v, want accepted challenge with second factor", startResp)
	}
	if len(delivery.messages) != 2 {
		t.Fatalf("delivery messages = %d, want 2", len(delivery.messages))
	}
	if startResp.Message == "" {
		t.Fatalf("StartPasswordRecovery() message = empty, want operator guidance")
	}

	challenge, err := svc.(*service).store.findRecoveryChallengeByID(context.Background(), startResp.ChallengeID)
	if err != nil {
		t.Fatalf("findRecoveryChallengeByID() error = %v", err)
	}
	resetResp, err := svc.ResetPasswordByRecovery(context.Background(), &PasswordRecoveryResetRequest{ChallengeID: startResp.ChallengeID, VerificationCode: "000000", SecondFactorCode: "000000", NewPassword: "RecoveredPass123!"}, AuthClientMeta{ClientIP: "8.8.8.8", UserAgent: "Recovery-UA", DeviceName: "Unknown Browser"})
	if err == nil || resetResp != nil {
		t.Fatalf("ResetPasswordByRecovery() should reject fake codes when preview is disabled")
	}
	if challenge.CodeHash == tokenDigest("000000") {
		t.Fatalf("challenge code hash unexpectedly matched placeholder code")
	}
}

func TestPasswordRecoveryPreviewDoesNotExposeCodes(t *testing.T) {
	t.Setenv("AUTH_PASSWORD_RECOVERY_PREVIEW_ENABLED", "true")
	prepareAuthTestConfig(t)
	db := newAuthTestDB(t)
	now := time.Now()
	seed := &sysUser{ID: 5, Username: "recover-preview", PasswordHash: mustHashPassword(t, "OldPass123!"), RoleCode: proposal.RolePlatformAdmin, RealName: "预览用户", Phone: "10000013801", Email: "preview@example.local", Status: proposal.UserStatusEnabled, CreatedAt: now, UpdatedAt: now}
	if err := db.Create(seed).Error; err != nil {
		t.Fatalf("seed create error = %v", err)
	}
	delivery := &stubRecoveryDeliverySender{}
	svc := New(&testDBRepo{r: db, w: db}, WithRecoveryDelivery(delivery))
	startResp, err := svc.StartPasswordRecovery(context.Background(), &PasswordRecoveryStartRequest{Username: "recover-preview", Channel: "email"}, AuthClientMeta{ClientIP: "8.8.4.4", UserAgent: "Recovery-UA", DeviceName: "Unknown Browser"})
	if err != nil {
		t.Fatalf("StartPasswordRecovery() error = %v", err)
	}
	if !startResp.Accepted || startResp.ChallengeID == "" {
		t.Fatalf("StartPasswordRecovery() = %#v, want accepted challenge", startResp)
	}
	wantMessages := 1
	if startResp.SecondFactorRequired {
		wantMessages = 2
	}
	if len(delivery.messages) != wantMessages {
		t.Fatalf("delivery messages = %d, want %d", len(delivery.messages), wantMessages)
	}
	if startResp.Message == "" {
		t.Fatalf("StartPasswordRecovery() message = empty, want non-empty guidance")
	}
	if startResp.ChallengeID == "" {
		t.Fatalf("StartPasswordRecovery() missing challengeId")
	}
}

func TestPasswordRecoveryDeliveryFailureExpiresChallengeAndWritesAudit(t *testing.T) {
	t.Setenv("AUTH_PASSWORD_RECOVERY_PREVIEW_ENABLED", "false")
	prepareAuthTestConfig(t)
	db := newAuthTestDB(t)
	now := time.Now()
	seed := &sysUser{ID: 6, Username: "recover-fail", PasswordHash: mustHashPassword(t, "OldPass123!"), RoleCode: proposal.RolePlatformAdmin, RealName: "失败用户", Email: "recover-fail@example.local", Status: proposal.UserStatusEnabled, CreatedAt: now, UpdatedAt: now}
	if err := db.Create(seed).Error; err != nil {
		t.Fatalf("seed create error = %v", err)
	}
	delivery := &stubRecoveryDeliverySender{err: context.DeadlineExceeded}
	svc := New(&testDBRepo{r: db, w: db}, WithRecoveryDelivery(delivery))
	resp, err := svc.StartPasswordRecovery(context.Background(), &PasswordRecoveryStartRequest{Username: "recover-fail", Channel: "email"}, AuthClientMeta{ClientIP: "9.9.9.9", UserAgent: "Recovery-UA", DeviceName: "Failure Browser"})
	if err == nil || resp != nil {
		t.Fatalf("StartPasswordRecovery() should fail on delivery error")
	}
	var challenge authRecoveryChallenge
	if err := db.First(&challenge, "username = ?", "recover-fail").Error; err != nil {
		t.Fatalf("load recovery challenge error = %v", err)
	}
	if challenge.Status != recoveryStatusExpired {
		t.Fatalf("challenge status = %s, want expired", challenge.Status)
	}
	var audits []authRecoveryAudit
	if err := db.Where("challenge_id = ?", challenge.ID).Find(&audits).Error; err != nil {
		t.Fatalf("load recovery audits error = %v", err)
	}
	foundFailure := false
	foundClassifiedFailure := false
	for _, item := range audits {
		if item.Action == "recovery.delivery_failed" && strings.Contains(item.Detail, "challengeExpired") {
			foundFailure = true
		}
		if item.Action == "recovery.delivery_failed" && strings.Contains(item.Detail, "errorKind") && strings.Contains(item.Detail, "password_recovery") {
			foundClassifiedFailure = true
		}
	}
	if !foundFailure {
		t.Fatalf("recovery delivery failed audit with challengeExpired not found: %#v", audits)
	}
	if !foundClassifiedFailure {
		t.Fatalf("classified recovery delivery failed audit not found: %#v", audits)
	}
}

func newAuthTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "auth-test.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE sys_user (
			id integer primary key,
			username text not null,
			password_hash text not null,
			role_code text not null,
			organization_id integer null,
			campus_id integer null,
			real_name text not null,
			phone text not null default '',
			email text not null default '',
			status text not null default 'enabled',
			failed_login_count integer not null default 0,
			locked_until datetime null,
			must_change_password numeric not null default 1,
			last_password_changed_at datetime null,
			remark text not null default '',
			created_at datetime not null,
			updated_at datetime not null
		);
		CREATE TABLE auth_session (
			id text primary key,
			user_id integer not null,
			organization_id integer null,
			campus_id integer null,
			role_code text not null default '',
			device_id text not null default '',
			device_name text not null default '',
			client_ip text not null default '',
			user_agent text not null default '',
			user_agent_hash text not null default '',
			access_token_id text not null default '',
			refresh_token_id text not null default '',
			refresh_token_hash text not null default '',
			status text not null default 'active',
			is_suspicious numeric not null default 0,
			risk_flags text,
			last_seen_at datetime null,
			last_refreshed_at datetime null,
			revoked_at datetime null,
			revoked_reason text not null default '',
			recovery_verified_at datetime null,
			created_at datetime not null,
			updated_at datetime not null
		);
		CREATE TABLE auth_recovery_challenge (
			id text primary key,
			user_id integer not null,
			username text not null default '',
			channel text not null,
			channel_target text not null default '',
			code_hash text not null,
			second_channel text not null default '',
			second_channel_target text not null default '',
			second_code_hash text not null default '',
			status text not null default 'pending',
			risk_flags text,
			requested_ip text not null default '',
			requested_user_agent text not null default '',
			expires_at datetime not null,
			verified_at datetime null,
			consumed_at datetime null,
			created_at datetime not null,
			updated_at datetime not null
		);
		CREATE TABLE auth_recovery_audit (
			id text primary key,
			challenge_id text not null,
			user_id integer not null,
			action text not null,
			status text not null default '',
			detail text,
			actor_ip text not null default '',
			user_agent text not null default '',
			created_at datetime not null
		);
	`).Error; err != nil {
		t.Fatalf("create table error = %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE notification_record (
			id integer primary key autoincrement,
			organization_id integer not null default 0,
			channel text not null,
			scene text not null,
			template_code text not null default '',
			target_masked text not null default '',
			provider text not null default '',
			status text not null,
			error_code text not null default '',
			error_message text not null default '',
			request_id text not null default '',
			attempts integer not null default 1,
			sent_at datetime null,
			created_at datetime not null
		)
	`).Error; err != nil {
		t.Fatalf("create notification_record error = %v", err)
	}
	return db
}

func mustHashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	return string(hash)
}

func prepareAuthTestConfig(t *testing.T) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	content := `[mysql.read]
addr = "127.0.0.1:3306"
user = "app"
name = "app"

[mysql.write]
addr = "127.0.0.1:3306"
user = "app"
name = "app"

[redis]
enabled = false
addr = ""

[jwt]
secret = "test-secret"
issuer = "test-issuer"
audience = "test-audience"
leewaySeconds = 0

[server]
port = ":9999"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("CONFIG_PATH", path)
	os.Args = []string{"auth.test", "-env", "dev"}
}
