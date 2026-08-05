package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

func GenRequestId() string {
	ts := time.Now().UnixMilli()
	randB := make([]byte, 4)
	rand.Read(randB)

	return fmt.Sprintf("%x-%s", ts, base64.RawStdEncoding.EncodeToString(randB))
}
