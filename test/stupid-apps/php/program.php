<?php

class Processor
{
    public function runProcess()
    {
        while (true) {
            echo "Processing...\n";
            usleep(200);
        }
    }
}

$process = new Processor();
$process->runProcess();

