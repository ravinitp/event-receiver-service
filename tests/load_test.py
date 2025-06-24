import requests
import time
import random
import string
import json
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime
import argparse
import logging

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

def generate_payload(size_bytes):
    """Generate a random payload of specified size in bytes."""
    body = ''.join(random.choices(string.ascii_letters + string.digits, k=size_bytes - 100))
    event = {
        "event_timestamp": datetime.utcnow().isoformat() + "Z",
        "body": body
    }
    payload = json.dumps(event)
    # Adjust payload size to be exact
    while len(payload.encode('utf-8')) < size_bytes:
        event["body"] += "a"
        payload = json.dumps(event)
    while len(payload.encode('utf-8')) > size_bytes:
        event["body"] = event["body"][:-1]
        payload = json.dumps(event)
    return payload

def send_request(endpoint, headers, payload, success_count, error_count):
    """Send a single POST request and update counters."""
    try:
        response = requests.post(endpoint, headers=headers, data=payload, timeout=5)
        if response.status_code == 200:
            success_count[0] += 1
            return response.elapsed.total_seconds()
        else:
            error_count[0] += 1
            logger.warning(f"Request failed with status {response.status_code}: {response.text}")
            return None
    except requests.RequestException as e:
        error_count[0] += 1
        logger.error(f"Request error: {str(e)}")
        return None

def run_load_test(endpoint, tier, payload_size_bytes, target_rps, duration_seconds):
    """Run load test with specified parameters."""
    logger.info(f"Starting load test: payload_size={payload_size_bytes/1024:.2f}KB, target_rps={target_rps}, duration={duration_seconds}s")

    headers = {"X-Customer-Tier": tier}
    payload = generate_payload(payload_size_bytes)

    success_count = [0]
    error_count = [0]
    response_times = []

    start_time = time.time()
    end_time = start_time + duration_seconds

    # Calculate number of workers needed to achieve target RPS
    workers = max(10, int(target_rps / 10))  # At least 10 workers, adjust based on RPS

    with ThreadPoolExecutor(max_workers=workers) as executor:
        while time.time() < end_time:
            futures = []
            batch_start = time.time()
            requests_to_send = int(target_rps / 10)  # Send in batches every 100ms

            for _ in range(requests_to_send):
                futures.append(executor.submit(
                    send_request, endpoint, headers, payload, success_count, error_count
                ))

            # Collect response times
            for future in futures:
                rt = future.result()
                if rt is not None:
                    response_times.append(rt)

            # Sleep to maintain target RPS
            elapsed = time.time() - batch_start
            if elapsed < 0.1:
                time.sleep(0.1 - elapsed)

    total_requests = success_count[0] + error_count[0]
    actual_rps = total_requests / duration_seconds
    avg_response_time = sum(response_times) / len(response_times) * 1000 if response_times else 0

    logger.info(f"Test completed:")
    logger.info(f"Total requests: {total_requests}")
    logger.info(f"Successful requests: {success_count[0]}")
    logger.info(f"Failed requests: {error_count[0]}")
    logger.info(f"Actual RPS: {actual_rps:.2f}")
    logger.info(f"Average response time: {avg_response_time:.2f}ms")

    return {
        "total_requests": total_requests,
        "success_count": success_count[0],
        "error_count": error_count[0],
        "actual_rps": actual_rps,
        "avg_response_time_ms": avg_response_time
    }

def main():
    parser = argparse.ArgumentParser(description="Load test for Event Receiver Service")
    parser.add_argument("--endpoint", default="http://localhost:8080/ingest", help="Endpoint URL")
    parser.add_argument("--tier", default="pro", choices=["free", "pro", "enterprise"], help="X-Customer-Tier header value")
    parser.add_argument("--min-size-kb", type=int, default=1, help="Minimum payload size in KB")
    parser.add_argument("--max-size-kb", type=int, default=10240, help="Maximum payload size in KB")
    parser.add_argument("--target-rps", type=int, default=100, help="Target requests per second")
    parser.add_argument("--duration", type=int, default=10, help="Test duration in seconds")

    args = parser.parse_args()

    # Test payload sizes: 1KB, 100KB, 1MB, 5MB, 10MB
    sizes_kb = [1, 100, 1000, 5000, 10240]
    sizes_kb = [size for size in sizes_kb if args.min_size_kb <= size <= args.max_size_kb]

    results = []
    for size_kb in sizes_kb:
        result = run_load_test(
            endpoint=args.endpoint,
            tier=args.tier,
            payload_size_bytes=size_kb * 1024,
            target_rps=args.target_rps,
            duration_seconds=args.duration
        )
        results.append((size_kb, result))

    # Print summary
    logger.info("\nTest Summary:")
    for size_kb, result in results:
        logger.info(f"Payload Size: {size_kb}KB")
        logger.info(f"  RPS: {result['actual_rps']:.2f}")
        logger.info(f"  Success: {result['success_count']}/{result['total_requests']}")
        logger.info(f"  Avg Response Time: {result['avg_response_time_ms']:.2f}ms")

if __name__ == "__main__":
    main()