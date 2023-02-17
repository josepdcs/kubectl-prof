import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

class Main {

  public static void main(String[] args) {
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
	public  void work(int n, long time, String function) {
    try {
            for (int i = 0; i < n; i++) {
                Thread.sleep(time);
                System.out.printf("Function: %s, Output: %d\n", function, i);
            }
    }
    catch (Exception e) {

            // catching the exception
            System.out.println(e);
        }

  }

  public  void fastFunction() {
    work(100, 50, "fastFunction");
  }

  public  void slowFunction() {
    work(100, 500, "slowFunction");
  }
}
