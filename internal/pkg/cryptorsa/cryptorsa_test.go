package cryptorsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func testKeyPair(t *testing.T) (string, string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey() error = %v", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPEM), string(privateKeyPEM)
}

func TestPublicKeyEncryptAndPrivateKeyDecrypt(t *testing.T) {
	publicKey, privateKey := testKeyPair(t)
	plaintext := "edu-schedule-system"

	ciphertext, err := PublicKeyEncrypt(publicKey, plaintext)
	if err != nil {
		t.Fatalf("PublicKeyEncrypt() error = %v", err)
	}
	decrypted, err := PrivateKeyDecrypt(privateKey, ciphertext)
	if err != nil {
		t.Fatalf("PrivateKeyDecrypt() error = %v", err)
	}
	if decrypted != plaintext {
		t.Fatalf("PrivateKeyDecrypt() = %q, want %q", decrypted, plaintext)
	}
}
