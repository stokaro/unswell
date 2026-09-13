#!/usr/bin/env bash
# Run inside the prepared container. Every invocation starts a new CLI process.
set -euo pipefail
cd /benchmark
mkdir -p results
uname -a >results/uname.txt
lscpu >results/lscpu.txt
dpkg-query -W >results/packages.txt
for name in cpu.max memory.max memory.swap.max; do
  cat "/sys/fs/cgroup/$name" >"results/$name"
done
for repetition in 1 2 3; do
  for corpus in synthetic pytest fastapi date-fns; do
    name=$corpus-$repetition
    args=()
    case "$corpus" in
      pytest) args=(--corpus /benchmark/sources/pytest --exclude testing/python/metafunc.py) ;;
      fastapi) args=(--corpus /benchmark/sources/fastapi --exclude 'docs/ja/**' --exclude 'docs/zh/**') ;;
      date-fns) args=(--corpus /benchmark/sources/date-fns --exclude 'pkgs/core/src/locale/**') ;;
    esac
    cat /proc/loadavg >"results/$name.load-before"
    cat /sys/fs/cgroup/memory.events >"results/$name.memory-before"
    bash scripts/measure-performance.sh --binary /benchmark/unswell --label "diabolocom-alpha3-$name" \
      --artifacts "/benchmark/results/$name" "${args[@]}" >"results/$name.log"
    cat /proc/loadavg >"results/$name.load-after"
    cat /sys/fs/cgroup/memory.events >"results/$name.memory-after"
    python3 scripts/summarize-performance.py "results/$name" >"results/$name.json"
    printf 'Recorded %s\n' "$name"
  done
done
