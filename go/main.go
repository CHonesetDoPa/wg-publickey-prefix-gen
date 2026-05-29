package main

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	mrand "math/rand"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var count uint64
var found uint64
var targetCount uint64
var x25519Curve = ecdh.X25519()
var encodedPubLen = base64.StdEncoding.EncodedLen(32)

func genKey() (string, string) {
	var priv [32]byte
	_, _ = rand.Read(priv[:])

	privKey, err := x25519Curve.NewPrivateKey(priv[:])
	if err != nil {
		return "", ""
	}
	pubBytes := privKey.PublicKey().Bytes()

	privStr := base64.StdEncoding.EncodeToString(priv[:])
	pubStr := base64.StdEncoding.EncodeToString(pubBytes)
	return privStr, pubStr
}

// genKeyFast generates a keypair using a non-crypto PRNG (per-worker) and
// writes the base64-encoded public key into pubEnc (must be sized using
// base64.StdEncoding.EncodedLen(32)). It returns the private scalar and the
// encoded length.
func genKeyFast(rng *mrand.Rand, pubEnc []byte) (priv [32]byte) {
	// fill private scalar with RNG bytes using Read for fewer calls
	_, _ = rng.Read(priv[:])

	privKey, err := x25519Curve.NewPrivateKey(priv[:])
	if err != nil {
		return priv
	}
	base64.StdEncoding.Encode(pubEnc[:encodedPubLen], privKey.PublicKey().Bytes())
	return priv
}

func worker(ctx context.Context, prefix string, id int, wg *sync.WaitGroup, cancel context.CancelFunc) {
	defer wg.Done()
	prefixLen := len(prefix)
	localCnt := 0

	// per-worker RNG and buffers (seeded once)
	var rng *mrand.Rand
	var pubEnc []byte
	prefixBytes := []byte(prefix)

	// seed RNG
	var seedBytes [8]byte
	_, _ = rand.Read(seedBytes[:])
	seed := int64(seedBytes[0])<<56 | int64(seedBytes[1])<<48 | int64(seedBytes[2])<<40 | int64(seedBytes[3])<<32 | int64(seedBytes[4])<<24 | int64(seedBytes[5])<<16 | int64(seedBytes[6])<<8 | int64(seedBytes[7])
	rng = mrand.New(mrand.NewSource(seed))
	pubEnc = make([]byte, base64.StdEncoding.EncodedLen(32))

	for {
		if ctx.Err() != nil {
			return
		}
		priv := genKeyFast(rng, pubEnc)
		encLen := encodedPubLen
		localCnt++

		if localCnt >= 1000 {
			atomic.AddUint64(&count, uint64(localCnt))
			localCnt = 0
		}

		if encLen >= prefixLen && bytes.Equal(pubEnc[:prefixLen], prefixBytes) {
			cur := atomic.AddUint64(&found, 1)

			if localCnt > 0 {
				atomic.AddUint64(&count, uint64(localCnt))
				localCnt = 0
			}

			fmt.Printf("\nFOUND #%d!\n", cur)
			fmt.Println("Thread:", id)
			// encode private key for output only
			privStr := base64.StdEncoding.EncodeToString(priv[:])
			pubStr := string(pubEnc[:encLen])
			fmt.Println("Private Key:", privStr)
			fmt.Println("Public Key :", pubStr)
			fmt.Println("Attempts so far:", atomic.LoadUint64(&count))

			if cur >= targetCount {
				// request cancellation of other workers
				cancel()
				return
			}
		}
	}
}

func monitor(ctx context.Context, start time.Time) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c := atomic.LoadUint64(&count)
			elapsed := time.Since(start).Seconds()
			speed := float64(c) / elapsed

			fmt.Printf("\rTried: %d | Speed: %.0f keys/sec | Elapsed: %.1fs",
				c, speed, elapsed)
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: wg_vanity <prefix> [threads] [count]")
		return
	}

	prefix := os.Args[1]

	threads := runtime.NumCPU()
	if len(os.Args) >= 3 {
		if v, err := strconv.Atoi(os.Args[2]); err == nil && v > 0 {
			threads = v
		} else {
			fmt.Println("Invalid threads value")
			return
		}
	}

	targetCount = 1
	if len(os.Args) >= 4 {
		if v, err := strconv.Atoi(os.Args[3]); err == nil && v > 0 {
			targetCount = uint64(v)
		} else {
			fmt.Println("Invalid count value")
			return
		}
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Println("Target prefix:", prefix)
	fmt.Println("Threads:", threads)
	fmt.Println("Need results:", targetCount)
	fmt.Println("CPU cores:", runtime.NumCPU())

	start := time.Now()

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	go func() {
		<-rootCtx.Done()
		if atomic.LoadUint64(&found) < targetCount {
			fmt.Println("\nReceived interrupt, shutting down...")
		}
	}()

	var wg sync.WaitGroup

	// Monitor
	go monitor(ctx, start)

	// Workers
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go worker(ctx, prefix, i, &wg, cancel)
	}

	wg.Wait()

	foundCount := atomic.LoadUint64(&found)
	totalTime := time.Since(start).Seconds()
	fmt.Printf("\nDone! Found %d keys in %.2f seconds\n", foundCount, totalTime)
}
