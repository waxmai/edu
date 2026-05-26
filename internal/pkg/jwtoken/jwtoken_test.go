package jwtoken

import (
	"fmt"
	"testing"
	"time"

	"edu-schedule-system/internal/proposal"
)

var secret = "i1yd" + "X9Rt" + "HyuJ" + "Trw7" + "frcu"

func TestSign(t *testing.T) {
	sessionUserInfo := proposal.SessionUserInfo{
		Id:       1001,
		UserName: "edu-schedule-system",
		NickName: "edu-schedule-system",
	}

	tokenString, err := New(secret).Sign(sessionUserInfo, 24*time.Hour)

	fmt.Println(tokenString, err)
	if err != nil {
		t.Error("sign error", err)
		return
	}

	t.Log(tokenString)
}

func TestParse(t *testing.T) {
	sessionUserInfo := proposal.SessionUserInfo{
		Id:       1001,
		UserName: "edu-schedule-system",
		NickName: "edu-schedule-system",
	}

	tokenString, err := New(secret).Sign(sessionUserInfo, time.Hour)
	if err != nil {
		t.Error("sign error", err)
		return
	}

	jwtInfo, err := New(secret).Parse(tokenString)
	if err != nil {
		t.Error("parse error", err)
		return
	}
	if jwtInfo.TokenID == "" {
		t.Fatal("token id should not be empty")
	}

	t.Log(jwtInfo)
}

func TestSignGeneratesDifferentTokensForSameUserWithinSameSecond(t *testing.T) {
	sessionUserInfo := proposal.SessionUserInfo{
		Id:       1001,
		UserName: "edu-schedule-system",
		NickName: "edu-schedule-system",
	}

	token1, err := New(secret).Sign(sessionUserInfo, time.Hour)
	if err != nil {
		t.Fatalf("first sign error: %v", err)
	}
	token2, err := New(secret).Sign(sessionUserInfo, time.Hour)
	if err != nil {
		t.Fatalf("second sign error: %v", err)
	}
	if token1 == token2 {
		t.Fatal("tokens should be different when signed twice")
	}
}
