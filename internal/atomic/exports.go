package atomic

import "sync/atomic"

// for import convenience, export things from sync/atomic that we use

type Int32 = atomic.Int32
type Int64 = atomic.Int64
type Uint32 = atomic.Uint32
type Uint64 = atomic.Uint64
