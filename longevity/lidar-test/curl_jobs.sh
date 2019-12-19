#! /bin/bash

while true; do seq 1 1000 | xargs -P 1000 -n1 bash -c 'curl http://lidar-drills.concourse-ci.org/api/v1/jobs'; sleep 1; done
