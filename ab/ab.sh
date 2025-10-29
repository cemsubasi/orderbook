#!/bin/sh

ab -n 10000 -c 100 -p buy_order.json -T 'application/json' http://localhost:8080/orders &
ab -n 10000 -c 100 -p sell_order.json -T 'application/json' http://localhost:8080/orders &
ab -n 10000 -c 100 -p trade.json -T 'application/json' http://localhost:8080/orders &