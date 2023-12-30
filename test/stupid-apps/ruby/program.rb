def fast_function
  loop do
    100.times do
        puts "Fast function"
        sleep(0.050)
    end
  end
end

def slow_function
  loop do
    100.times do
        puts "Slow function"
        sleep(0.500)
    end
  end
end

def process1
  t1 = Thread.new { fast_function }
  t2 = Thread.new { slow_function }
  t1.join
  t2.join
end

def process2
  t1 = Thread.new { fast_function }
  t2 = Thread.new { slow_function }
  t1.join
  t2.join
end

process1_pid = fork { process1 }
process2_pid = fork { process2 }

Process.wait(process1_pid)
Process.wait(process2_pid)
