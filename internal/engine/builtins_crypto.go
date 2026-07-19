package engine

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"hash"
	"hash/crc32"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

// digestAlgo is a digest algorithm name (case-insensitive). The weak digests
// (sha1, md5) are deliberately absent: the fleet gate blocklists them, and a
// tamper-evidence family must not ship broken primitives (ADR 0011 §1).
type digestAlgo string

// algoDefault is the algorithm DIGEST and HMAC use when none is named.
const algoDefault digestAlgo = "sha256"

// hashNamed is the constructor for a named HMAC-capable algorithm; ok is
// false for an unknown name.
func hashNamed(algo digestAlgo) (func() hash.Hash, boolResult) {
	switch strings.ToLower(string(algo)) {
	case string(algoDefault):
		return sha256.New, true
	case "sha512":
		return sha512.New, true
	default:
		return nil, false
	}
}

// writeAll feeds text to a hash; hash.Hash.Write never returns an error (its
// documented contract), so the counts are discarded.
func writeAll(h hash.Hash, text textVal) {
	_, _ = h.Write([]byte(text))
}

// hexSum is the lowercase hex digest of text under an algorithm constructor.
func hexSum(newHash func() hash.Hash, text textVal) Value {
	h := newHash()
	writeAll(h, text)
	return stringValue(textVal(hex.EncodeToString(h.Sum(nil))))
}

// fnHash lifts a named digest into an eager impl over the argument's
// canonical text form (ADR 0011 §4: lowercase hex over UTF-8 bytes).
func fnHash(algo digestAlgo) func(args []Value) Value {
	return func(args []Value) Value { return digestText(algo, textVal(argText(args, 0))) }
}

// fnCrc32 is the IEEE CRC-32 of the argument's text form, as 8 hex digits —
// an explicitly non-cryptographic checksum.
func fnCrc32(args []Value) Value {
	return stringValue(textVal(crcHex(textVal(argText(args, 0)))))
}

// crcHex is the zero-padded 8-hex-digit IEEE CRC-32 of text.
func crcHex(text textVal) string {
	digits := strconv.FormatUint(uint64(crc32.ChecksumIEEE([]byte(text))), 16)
	return strings.Repeat("0", 8-len(digits)) + digits
}

// digestText hashes text under a named algorithm; crc32 joins the named set
// here (DIGEST accepts it; HMAC does not). An unknown name is #VALUE!.
func digestText(algo digestAlgo, text textVal) Value {
	if strings.ToLower(string(algo)) == "crc32" {
		return stringValue(textVal(crcHex(text)))
	}
	newHash, ok := hashNamed(algo)
	if !ok {
		return errorValue(ErrValue)
	}
	return hexSum(newHash, text)
}

// fnHmac is HMAC(key, text, [algo]) — the keyed digest in lowercase hex,
// SHA-256 unless another HMAC-capable algorithm is named (ADR 0011 §4).
func fnHmac(args []Value) Value {
	algo := algoDefault
	if len(args) == 3 {
		algo = digestAlgo(argText(args, 2))
	}
	newHash, ok := hashNamed(algo)
	if !ok {
		return errorValue(ErrValue)
	}
	mac := hmac.New(newHash, []byte(argText(args, 0)))
	writeAll(mac, textVal(argText(args, 1)))
	return stringValue(textVal(hex.EncodeToString(mac.Sum(nil))))
}

// fnBase64 encodes the argument's text form with the standard padded alphabet.
func fnBase64(args []Value) Value {
	return stringValue(textVal(base64.StdEncoding.EncodeToString([]byte(argText(args, 0)))))
}

// fnUnbase64 reverses fnBase64; malformed input is #VALUE!. The name follows
// Hive/SQL (UNBASE64) because the grammar admits digits only at a name's tail
// — `base64decode` is unlexable, and the grammar is untouched (ADR 0011).
func fnUnbase64(args []Value) Value {
	return decodedText(base64.StdEncoding.DecodeString(argText(args, 0)))
}

// fnHex encodes the argument's text form as lowercase hex.
func fnHex(args []Value) Value {
	return stringValue(textVal(hex.EncodeToString([]byte(argText(args, 0)))))
}

// fnUnhex reverses fnHex; malformed input is #VALUE!. Named for symmetry with
// fnUnbase64 (Hive/SQL UNHEX).
func fnUnhex(args []Value) Value {
	return decodedText(hex.DecodeString(argText(args, 0)))
}

// decodedText lifts a byte decoding into a text value: malformed input, or a
// payload that is not valid UTF-8, is #VALUE! (cells hold text, never raw
// bytes).
func decodedText(raw []byte, err error) Value {
	if err != nil || !utf8.Valid(raw) {
		return errorValue(ErrValue)
	}
	return stringValue(textVal(string(raw)))
}

// isDigest reports whether name is a range-digesting builtin.
func isDigest(name funcName) boolResult {
	return name == "digest" || name == "verify"
}

// evalDigest dispatches DIGEST and VERIFY, which read a range's 2-D shape to
// serialize it canonically (ADR 0011 §3). ok is false for any other name.
func (r resolver) evalDigest(name funcName, args []tsvt.Expr) (Value, boolResult) {
	switch name {
	case "digest":
		return r.digestRange(args), true
	case "verify":
		return r.verifyRange(args), true
	default:
		return Value{}, false
	}
}

// canonicalTSV serializes a range's computed values canonically: each row's
// cell texts TAB-joined, every row LF-terminated — the byte-pinned message of
// DIGEST and VERIFY (ADR 0011 §3). An error cell propagates (strict rule).
func canonicalTSV(m [][]Value) (textVal, Value) {
	rows := make([]string, len(m))
	for i, row := range m {
		texts := make([]string, len(row))
		for c, v := range row {
			if v.isError() {
				return "", v
			}
			texts[c] = v.String()
		}
		rows[i] = strings.Join(texts, "\t")
	}
	return textVal(strings.Join(rows, "\n") + "\n"), Value{}
}

// digestRange evaluates DIGEST(range, [algo]) over the range's canonical
// serialization — the cell that makes a sheet tamper-evident.
func (r resolver) digestRange(args []tsvt.Expr) Value {
	if len(args) < 1 || len(args) > 2 {
		return errorValue(ErrValue)
	}
	algo := algoDefault
	if len(args) == 2 {
		named, bad := r.algoArg(args[1])
		if bad.isError() {
			return bad
		}
		algo = named
	}
	msg, bad := canonicalTSV(r.argMatrix(args[0]))
	if bad.isError() {
		return bad
	}
	return digestText(algo, msg)
}

// algoArg reads an algorithm-name argument; an error value propagates.
func (r resolver) algoArg(arg tsvt.Expr) (digestAlgo, Value) {
	v := r.argScalar(arg)
	if v.isError() {
		return "", v
	}
	return digestAlgo(v.String()), Value{}
}

// verifyRange evaluates VERIFY(signature, pubkey, range): TRUE iff the
// hex-encoded Ed25519 signature verifies over the range's canonical
// serialization under the hex-encoded public key; an undecodable or
// wrong-length signature or key is #VALUE! (ADR 0011 §4).
func (r resolver) verifyRange(args []tsvt.Expr) Value {
	if len(args) != 3 {
		return errorValue(ErrValue)
	}
	sig, sigBad := hexArg(r.argScalar(args[0]), byteLen(ed25519.SignatureSize))
	if sigBad.isError() {
		return sigBad
	}
	key, keyBad := hexArg(r.argScalar(args[1]), byteLen(ed25519.PublicKeySize))
	if keyBad.isError() {
		return keyBad
	}
	msg, bad := canonicalTSV(r.argMatrix(args[2]))
	if bad.isError() {
		return bad
	}
	return boolValue(boolResult(ed25519.Verify(ed25519.PublicKey(key), []byte(msg), sig)))
}

// byteLen is an exact byte length a decoded argument must have.
type byteLen int

// hexArg decodes a hex-encoded argument of an exact byte size; an error value
// propagates and anything else malformed is #VALUE!.
func hexArg(v Value, size byteLen) ([]byte, Value) {
	if v.isError() {
		return nil, v
	}
	raw, err := hex.DecodeString(v.String())
	if err != nil || byteLen(len(raw)) != size {
		return nil, errorValue(ErrValue)
	}
	return raw, Value{}
}
