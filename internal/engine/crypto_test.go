package engine_test

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

func TestCrypto_TextDigests(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		`sha256("")`:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		`sha256("abc")`: "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
		`sha512("abc")`: "ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a" +
			"2192992a274fc1a836ba3c23a3feebbd454d4423643ce80e2a9ac94fa54ca49f",
		`crc32("abc")`: "352441c2",
		`crc32("c")`:   "06b9df6f", // zero-padded to 8 digits
	}
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, want, formula1(t, expr))
		})
	}
}

func TestCrypto_DigestSerializesCanonically(t *testing.T) {
	t.Parallel()

	// A single cell "x" serializes as "x\n".
	one := compute(t, "x\t=digest(A1)\n")
	assert.Equal(t, "73cb3858a687a8494ca3323053016282f3dad39d42cf62ca4e79dda2aac7d9ac", cellAt(t, one, 0, 1))

	// A 2×2 block serializes as "1\t2\n3\t4\n" — TAB-joined cells, every row
	// LF-terminated (ADR 0011 §3).
	block := compute(t, "1\t2\t=digest(A1:B2)\n3\t4\n")
	assert.Equal(t, "59e75fb0f9f5f4e46f6c93e63ac1a4f19fa43af72fdbcd3f8789a4d60c4b8875", cellAt(t, block, 0, 2))
}

func TestCrypto_DigestNamedAlgorithms(t *testing.T) {
	t.Parallel()

	row := compute(t, "1\t2\t=digest(A1:B1, \"sha512\")\t=digest(A1:B1, \"SHA256\")\n")
	assert.Equal(t, "598d558443f8cbea7a9806d5f55013e8638fabb50e50e57285278ed2b92e46c6"+
		"419e064c0fb00a975884776907522b59daaea9a5cf35f68502c591eeebcc8eaf", cellAt(t, row, 0, 2))
	// Names are case-insensitive; the default is sha256.
	assert.Equal(t, cellAt(t, compute(t, "1\t2\t=digest(A1:B1)\n"), 0, 2), cellAt(t, row, 0, 3))

	crc := compute(t, "x\t=digest(A1, \"crc32\")\n")
	assert.Equal(t, "46ea081f", cellAt(t, crc, 0, 1))
}

func TestCrypto_DigestErrors(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		src  string
		want engine.ErrorValue
	}{
		"unknown algorithm":       {"x\t=digest(A1, \"nope\")\n", engine.ErrValue},
		"weak digests are absent": {"x\t=digest(A1, \"md5\")\n", engine.ErrValue},
		"algorithm error":         {"x\t=digest(A1, Z99)\n", engine.ErrRef},
		"no arguments":            {"x\t=digest()\n", engine.ErrValue},
		"too many arguments":      {"x\t=digest(A1, \"sha256\", 1)\n", engine.ErrValue},
		"error cell propagates":   {"=1/0\t=digest(A1)\n", engine.ErrDiv},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, string(tc.want), cellAt(t, compute(t, tc.src), 0, 1))
		})
	}
}

func TestCrypto_Hmac(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "9c196e32dc0175f86f4b1cb89289d6619de6bee699e4c378e68309ed97a1a6ab",
		formula1(t, `hmac("key", "abc")`)) // sha256 default
	assert.Equal(t, "3926a207c8c42b0c41792cbd3e1a1aaaf5f7a25704f62dfc939c4987dd7ce060"+
		"009c5bb1c2447355b3216f10b537e9afa7b64a4e5391b0d631172d07939e087a",
		formula1(t, `hmac("key", "abc", "sha512")`))
	// CRC-32 is not HMAC-capable; unknown names are #VALUE!.
	assert.Equal(t, string(engine.ErrValue), formula1(t, `hmac("key", "abc", "crc32")`))
}

func TestCrypto_Codecs(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		`base64("abc")`:    "YWJj",
		`base64("")`:       "",
		`unbase64("YWJj")`: "abc",
		`hex("abc")`:       "616263",
		`unhex("616263")`:  "abc",
		`unbase64("!!")`:   string(engine.ErrValue), // malformed
		`unbase64("/w==")`: string(engine.ErrValue), // 0xFF is not UTF-8
		`unhex("zz")`:      string(engine.ErrValue), // malformed
		`unhex("ff")`:      string(engine.ErrValue), // 0xFF is not UTF-8
	}
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, want, formula1(t, expr))
		})
	}
}

func TestCrypto_VerifyEd25519(t *testing.T) {
	t.Parallel()

	private := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{7}, ed25519.SeedSize))
	public := hex.EncodeToString(private.Public().(ed25519.PublicKey))
	signature := hex.EncodeToString(ed25519.Sign(private, []byte("1\t2\n")))

	good := compute(t, "1\t2\n=verify(\""+signature+"\", \""+public+"\", A1:B1)\n")
	assert.Equal(t, "TRUE", cellAt(t, good, 1, 0))

	// Any change to the signed range flips the verification — tamper-evidence.
	tampered := compute(t, "1\t9\n=verify(\""+signature+"\", \""+public+"\", A1:B1)\n")
	assert.Equal(t, "FALSE", cellAt(t, tampered, 1, 0))
}

func TestCrypto_VerifyErrors(t *testing.T) {
	t.Parallel()

	private := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{7}, ed25519.SeedSize))
	public := hex.EncodeToString(private.Public().(ed25519.PublicKey))
	signature := hex.EncodeToString(ed25519.Sign(private, []byte("1\n")))

	cases := map[string]struct {
		src  string
		want engine.ErrorValue
	}{
		"wrong arity":          {"1\n=verify(A1)\n", engine.ErrValue},
		"undecodable sig":      {"1\n=verify(\"zz\", \"" + public + "\", A1)\n", engine.ErrValue},
		"wrong-length sig":     {"1\n=verify(\"aa\", \"" + public + "\", A1)\n", engine.ErrValue},
		"wrong-length key":     {"1\n=verify(\"" + signature + "\", \"aa\", A1)\n", engine.ErrValue},
		"sig error propagates": {"1\n=verify(Z99, \"" + public + "\", A1)\n", engine.ErrRef},
		"range error":          {"=1/0\n=verify(\"" + signature + "\", \"" + public + "\", A1)\n", engine.ErrDiv},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, string(tc.want), cellAt(t, compute(t, tc.src), 1, 0))
		})
	}
}
