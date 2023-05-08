#!/usr/bin/env bash

(source .venv/bin/activate && uvicorn main:app --port 9000) &

(PORT=9100 ./node ./main.js) &

(PORT=8000 ./proxy) &

wait

exit $?
