using System;
using System.Threading;
using System.Threading.Tasks;

static void FastFunction()
{
    while (true)
    {
        for (int i = 0; i < 100; i++)
        {
            Console.WriteLine($"fast_function: {i}");
            Thread.Sleep(25);
        }
    }
}

static void SlowFunction()
{
    while (true)
    {
        for (int i = 0; i < 100; i++)
        {
            Console.WriteLine($"slow_function: {i}");
            Thread.Sleep(150);
        }
    }
}

await Task.WhenAll(
    Task.Run(() => FastFunction()),
    Task.Run(() => SlowFunction())
);
