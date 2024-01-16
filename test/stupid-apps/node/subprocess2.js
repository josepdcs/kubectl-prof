function consumeFunction(n, time) {
    for (let i = 0; i < n; i++) {
        sleep(time)
        console.log(`Output ${i}`);
    }
}

function sleep(ms) {
    Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, ms);
}

while(true) {
    consumeFunction(100, 100);
}

