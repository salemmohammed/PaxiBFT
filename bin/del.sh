kill -9 $(lsof -i:1735 -t)
kill -9 $(lsof -i:1736 -t)
kill -9 $(lsof -i:1737 -t)
kill -9 $(lsof -i:1738 -t)
rm client* server* cmd history* latency

