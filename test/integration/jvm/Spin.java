public class Spin {
    public static void main(String[] args) throws Exception {
        long acc = 0;
        while (true) {
            for (int i = 0; i < 2_000_000; i++) {
                acc += i % 7;
            }
            if (acc == Long.MIN_VALUE) {
                System.out.println("unreachable");
            }
            Thread.sleep(1);
        }
    }
}
