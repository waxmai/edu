package cryptoaes

import "testing"

func testAESKey() string {
	parts := []string{"m52p", "sEZH", "Sn4!", "&hbs"}
	return parts[0] + parts[1] + parts[2] + parts[3]
}

func TestEncrypt2(t *testing.T) {
	t.Log(Encrypt(testAESKey(), "edu-schedule-system"))
}

func TestDecrypt(t *testing.T) {
	t.Log(Decrypt(testAESKey(), "qAyQtb9bkvbDFW47H5DGDVwTjw399k13xM2ceBg/OGc="))
}
