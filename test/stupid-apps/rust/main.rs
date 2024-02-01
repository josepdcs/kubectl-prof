use std::thread;
use std::time::Duration;

fn fast_function() {
    loop {
        for _ in 0..100 {
            println!("fast_function");
            thread::sleep(Duration::from_millis(100));
        }
    }
}

fn slow_function() {
   loop {
           for _ in 0..100 {
               println!("slow_function");
               thread::sleep(Duration::from_millis(500));
           }
       }
}

fn main() {
    let fast_function_handle = thread::spawn(|| fast_function());
    let slow_function_handle = thread::spawn(|| slow_function());

    fast_function_handle.join().unwrap();
    slow_function_handle.join().unwrap();

}
