#!/bin/bash
listen=$1

if [[ -z $listen ]];
then
	echo 'usage: ./stop.sh <listen>'
else
	ps aux | grep $1 | grep -v grep | awk '{print $2}' | xargs kill
fi
