package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

// GenerateUUID returns a random UUID v4 using crypto/rand.
func GenerateUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("crypto/rand failure: " + err.Error())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:]),
	)
}

const appCodeCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

const appCodeSuffixLen = 6

func GenerateApplicationCode() string {
	return GenerateApplicationCodeAt(time.Now())
}

func GenerateApplicationCodeAt(t time.Time) string {
	suffix := make([]byte, appCodeSuffixLen)
	max := big.NewInt(int64(len(appCodeCharset)))
	for i := 0; i < appCodeSuffixLen; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			// crypto/rand failing is an unrecoverable environment problem
			panic("crypto/rand failure: " + err.Error())
		}
		suffix[i] = appCodeCharset[n.Int64()]
	}
	return "APP-" + t.Format("20060102") + "-" + string(suffix)
}
