import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

public class Main {

    public static void main(String[] args) {
        System.out.println("Starting Non-Root JVM Stupid App...");
        ExecutorService executorService = Executors.newFixedThreadPool(2);
        executorService.submit(new FastRunner());
        executorService.submit(new SlowRunner());
    }
}

class FastRunner implements Runnable {

    @Override
    public void run() {
        Worker w = new Worker();
        while (true) {
            w.fastFunction();
        }
    }
}

class SlowRunner implements Runnable {

    @Override
    public void run() {
        Worker w = new Worker();
        while (true) {
            w.slowFunction();
        }
    }
}

class Worker {
    public void work(int n, long time, String function) {
        try {
            long acc = 0;
            for (int i = 0; i < n; i++) {
                for (int j = 0; j < 50_000; j++) {
                    acc += (i + j) % 7;
                }
                Thread.sleep(time);
                if (acc == Long.MIN_VALUE) {
                    System.out.println("unreachable");
                }
            }
            System.out.printf("Function: %s completed iteration cycle\n", function);
        } catch (Exception e) {
            System.out.println("Worker interrupted: " + e.getMessage());
        }
    }

    public void fastFunction() {
        work(100, 50, "fastFunction");
    }

    public void slowFunction() {
        work(100, 200, "slowFunction");
    }
}
