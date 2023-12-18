#!/bin/bash

ps $1 | awk -v p='COMMAND' 'NR==1 {n=index($0, p); next} {print substr($0, n)}'