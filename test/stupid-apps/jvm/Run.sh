#!/bin/bash

# Start the first process
/opt/app/Run2.sh &

# Start the second process
/opt/app/Run2.sh &

# Wait for any process to exit
wait -n

# Exit with status of process that exited first
exit $?

