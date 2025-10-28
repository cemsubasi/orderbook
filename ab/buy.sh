#!/bin/sh

ab -n 10000 -c 100 -p buy.json -T 'application/json' http://localhost:8080/orders
