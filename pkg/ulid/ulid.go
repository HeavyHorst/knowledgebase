package ulid

import (
	"math/rand"
	"sync"
	"time"

	"github.com/oklog/ulid"
)

var randPool = sync.Pool{
	New: func() interface{} {
		return rand.New(rand.NewSource(time.Now().UnixNano()))
	},
}

func GetULID() string {
	entropy := randPool.Get().(*rand.Rand)
	defer randPool.Put(entropy)
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}
