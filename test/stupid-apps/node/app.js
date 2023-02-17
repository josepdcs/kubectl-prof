while (true) {
    consumeFunction(100, 250);
}

function consumeFunction(n, time) {
    for (let i = 0; i < n; i++) {
        sleep(time)
        console.log(`Output ${i}`);
    }
}

function sleep(milliseconds) {
    const date = Date.now();
    let currentDate = null;
    do {
        currentDate = Date.now();
    } while (currentDate - date < milliseconds);
}
