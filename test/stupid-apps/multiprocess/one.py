#!/usr/bin/python
import asyncio


async def work(n, time, function):
    for i in range(n):
        print(f"Function: {function}, Output: {i}")
        await asyncio.sleep(time / 1000)


async def fast_function():
    while True:
        await work(100, 50, "fast_function")


async def slow_function():
    while True:
        await work(100, 500, "slow_function")


async def launch():
    tasks = [asyncio.ensure_future(coro()) for coro in (fast_function, slow_function)]
    await asyncio.wait(tasks)


if __name__ == "__main__":
    loop = asyncio.get_event_loop()
    loop.run_until_complete(launch())
