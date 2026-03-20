using System;
using System.Threading;
using System.Threading.Tasks;

static void FastFunction(int n)
{
    for (int i = 0; i < n; i++)
    {
        Console.WriteLine($"fast_function: {i}");
        Thread.Sleep(25);
    }
}

static void SlowFunction(int n)
{
    for (int i = 0; i < n; i++)
    {
        Console.WriteLine($"slow_function: {i}");
        Thread.Sleep(150);
    }
}

static async Task RunAsync()
{
    while (true)
    {
        await Task.Run(() => FastFunction(100));
        await Task.Run(() => SlowFunction(100));
    }
}

await RunAsync();
