<?php

function fast_function(int $n): void {
    for ($i = 0; $i < $n; $i++) {
        echo "fast_function: $i\n";
        usleep(25000);
    }
}

function slow_function(int $n): void {
    for ($i = 0; $i < $n; $i++) {
        echo "slow_function: $i\n";
        usleep(150000);
    }
}

function work(): void {
    while (true) {
        fast_function(100);
        slow_function(100);
    }
}

work();
