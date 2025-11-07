const TARGET = "http://localhost:8080/";
const CONCURRENCY = 20;  // number of concurrent requests
const REQUESTS = 100;     // total number of requests

async function delay(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function makeRequest(i) {
  const start = Date.now();
  try {
    const res = await fetch(TARGET, {
      headers: {
        "X-App-ID": "test-app1",
      },
    });
    const text = await res.text();
    const elapsed = Date.now() - start;
    console.log(
      `#${i.toString().padStart(2)} ✅ ${res.status} | ${elapsed}ms | ${text.trim()}`
    );
  } catch (err) {
    console.error(`#${i.toString().padStart(2)} ❌ Error:`, err.message);
  }
}

async function run() {
  console.log(`🚀 Starting ${REQUESTS} requests with concurrency=${CONCURRENCY}`);
  const queue = Array.from({ length: REQUESTS }, (_, i) => i + 1);

  while (queue.length > 0) {
    const batch = queue.splice(0, CONCURRENCY);
    await Promise.all(batch.map(makeRequest));
    await delay(50); // small pause between bursts
  }

  console.log("✅ Test complete!");
}

run();
