#!/usr/bin/env python3
import argparse
import base64
import multiprocessing as mp
import time
import signal
import sys
import os
from queue import Empty

try:
    from cryptography.hazmat.primitives.asymmetric import x25519
    from cryptography.hazmat.primitives import serialization
except Exception as e:
    print("Missing dependency: cryptography is required. Install with: pip install cryptography")
    sys.exit(1)


parser = argparse.ArgumentParser(description="Local X25519 WireGuard vanity key generator")
parser.add_argument("prefix", help="public key base64 prefix to search for")
parser.add_argument("threads", type=int, nargs="?", default=None, help="number of processes (default: CPU cores)")
parser.add_argument("count", type=int, nargs="?", default=1, help="number of results to find (default: 1)")
args = parser.parse_args()

prefix = args.prefix
PROCESS_COUNT = args.threads if args.threads and args.threads > 0 else os.cpu_count() or 1

# Shared objects will be created in __main__ when using multiprocessing


def gen_x25519_keys():
    priv = x25519.X25519PrivateKey.generate()
    priv_bytes = priv.private_bytes(
        encoding=serialization.Encoding.Raw,
        format=serialization.PrivateFormat.Raw,
        encryption_algorithm=serialization.NoEncryption(),
    )
    pub = priv.public_key()
    pub_bytes = pub.public_bytes(
        encoding=serialization.Encoding.Raw,
        format=serialization.PublicFormat.Raw,
    )
    priv_b64 = base64.b64encode(priv_bytes).decode().strip()
    pub_b64 = base64.b64encode(pub_bytes).decode().strip()
    return priv_b64, pub_b64


def worker(tid, prefix, stop_event, lock, count, found, start_time, target_count, results_q, flush_interval=1000):
    # Ignore SIGINT in worker processes; main process handles shutdown
    try:
        signal.signal(signal.SIGINT, signal.SIG_IGN)
    except Exception:
        pass
    local_count = 0
    while not stop_event.is_set():
        try:
            priv, pub = gen_x25519_keys()
        except Exception:
            continue

        local_count += 1

        # Periodically flush local counts to shared counter to reduce lock contention
        if local_count >= flush_interval:
            with lock:
                count.value += local_count
            local_count = 0

        if pub.startswith(prefix):
            # flush remaining local_count
            with lock:
                count.value += local_count
                local_count = 0
                found.value += 1
                cur = found.value

            # send result to main process for printing
            try:
                results_q.put_nowait((tid, priv, pub, cur))
            except Exception:
                pass

            if cur >= target_count:
                stop_event.set()
                break

    # Flush remaining local attempts when worker exits to keep final stats accurate
    if local_count:
        with lock:
            count.value += local_count


if __name__ == "__main__":
    stop_event = mp.Event()
    lock = mp.Lock()
    count = mp.Value('Q', 0)
    found = mp.Value('Q', 0)
    target_count = args.count if args.count and args.count > 0 else 1
    start_time = mp.Value('d', time.time())

    processes = []
    results_q = mp.Queue()

    def monitor(stop_event, count, start_time, found):
        # Ignore SIGINT in monitor process so main handles Ctrl+C
        try:
            signal.signal(signal.SIGINT, signal.SIG_IGN)
        except Exception:
            pass
        # Print status every second like Go monitor
        while not stop_event.is_set():
            time.sleep(1)
            with lock:
                tried = count.value
                found_local = found.value
            elapsed = time.time() - start_time.value
            rate = tried / elapsed if elapsed > 0 else 0
            # Use ANSI escape to clear to end of line to avoid leftover characters
            # Match Go monitor format (no 'Found' field)
            line = f"Tried: {tried} | Speed: {rate:.0f} keys/sec | Elapsed: {elapsed:.1f}s"
            print(f"\r{line}\033[K", end="", flush=True)

    print(f"Target prefix: {prefix}")
    print(f"Threads: {PROCESS_COUNT}")
    print(f"Need results: {target_count}")
    print(f"CPU cores: {os.cpu_count()}")

    # start monitor
    mon = mp.Process(target=monitor, args=(stop_event, count, start_time, found))
    mon.start()

    # Install signal handlers in main process to gracefully shutdown on Ctrl+C / SIGTERM
    def _shutdown(signum, frame):
        print("\nReceived interrupt, shutting down...")
        stop_event.set()

    try:
        signal.signal(signal.SIGINT, _shutdown)
    except Exception:
        pass
    try:
        signal.signal(signal.SIGTERM, _shutdown)
    except Exception:
        pass

    for i in range(PROCESS_COUNT):
        p = mp.Process(target=worker, args=(i, prefix, stop_event, lock, count, found, start_time, target_count, results_q))
        p.start()
        processes.append(p)

    # main loop: print found items as they arrive
    try:
        while True:
            try:
                tid, priv, pub, cur = results_q.get(timeout=0.5)
                print(f"\nFOUND #{cur}!")
                print("Thread:", tid)
                print("Private Key:", priv)
                print("Public Key :", pub)
                with lock:
                    print("Attempts so far:", count.value)
                if found.value >= target_count:
                    break
            except Empty:
                if stop_event.is_set():
                    break
                # otherwise continue waiting
                continue
    finally:
        # ensure all processes terminate
        stop_event.set()
        for p in processes:
            p.join()
        mon.join()

    totalTime = time.time() - start_time.value
    with lock:
        final_found = found.value
    print(f"\nDone! Found {final_found} keys in {totalTime:.2f} seconds")
