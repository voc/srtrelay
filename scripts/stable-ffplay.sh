#!/bin/sh
# srt-live-transmit wrapper which actually manages to reconnect a failed srt stream
url=$1
pid=$$
fifo="/tmp/srt.${pid}"

trap 'rm "${fifo}"' EXIT

while true; do
    srt-live-transmit -a no "${url}" file://con > ${fifo} &
    p1=$!
    ffplay -v warning -hide_banner -nostats ${fifo} &
    p2=$!

    # check whether we still have 2 running jobs
    num_jobs="$(jobs -p | wc -l)"
    while [ $num_jobs -ge 2 ]; do
        sleep 1
        num_jobs="$(jobs -p | wc -l)"
    done

    kill ${p1} ${p2}
    sleep 1
done
