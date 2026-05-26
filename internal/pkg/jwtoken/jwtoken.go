package jwtoken

import (
	"errors"
	"time"

	"edu-schedule-system/internal/pkg/idgen"
	"edu-schedule-system/internal/proposal"

	"github.com/golang-jwt/jwt/v5"
)

var _ Token = (*token)(nil)

type Token interface {
	i()
	Sign(jwtInfo proposal.SessionUserInfo, expireDuration time.Duration) (tokenString string, err error)
	Parse(tokenString string) (*claims, error)
}

type token struct {
	secret   string
	issuer   string
	audience string
	leeway   time.Duration
}

type claims struct {
	proposal.SessionUserInfo
	jwt.RegisteredClaims
}

type Option func(*token)

func WithIssuer(issuer string) Option {
	return func(t *token) {
		t.issuer = issuer
	}
}

func WithAudience(audience string) Option {
	return func(t *token) {
		t.audience = audience
	}
}

func WithLeeway(leeway time.Duration) Option {
	return func(t *token) {
		t.leeway = leeway
	}
}

func New(secret string, opts ...Option) Token {
	t := &token{
		secret: secret,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (t *token) i() {}

func (t *token) Sign(sessionUserInfo proposal.SessionUserInfo, expireDuration time.Duration) (tokenString string, err error) {
	now := time.Now()
	if sessionUserInfo.TokenID == "" {
		sessionUserInfo.TokenID = idgen.GenerateUniqueID()
	}
	claims := claims{
		sessionUserInfo,
		jwt.RegisteredClaims{
			ID:        sessionUserInfo.TokenID,
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expireDuration)),
			Issuer:    t.issuer,
		},
	}
	if t.audience != "" {
		claims.Audience = jwt.ClaimStrings{t.audience}
	}

	tokenString, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(t.secret))
	return
}

func (t *token) Parse(tokenString string) (*claims, error) {
	options := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	}
	if t.issuer != "" {
		options = append(options, jwt.WithIssuer(t.issuer))
	}
	if t.audience != "" {
		options = append(options, jwt.WithAudience(t.audience))
	}
	if t.leeway > 0 {
		options = append(options, jwt.WithLeeway(t.leeway))
	}

	tokenClaims, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(t.secret), nil
	}, options...)

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*claims); ok && tokenClaims.Valid {
			if claims.SessionUserInfo.TokenID == "" {
				claims.SessionUserInfo.TokenID = claims.ID
			}
			return claims, nil
		}
	}

	return nil, err
}
