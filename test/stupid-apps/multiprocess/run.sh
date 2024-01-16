#!/bin/bash

# Start the first process
python /app/one.py &

# Start the second process
python /app/two.py &

# Wait for any process to exit
wait -n

# Exit with status of process that exited first
exit $?