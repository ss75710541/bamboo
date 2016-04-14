docker run -it \
    --net=host \
    -v /data/haproxy:/etc/haproxy \
    -e MARATHON_ENDPOINT=http://10.3.10.83:5098 \
    -e BAMBOO_ENDPOINT=http://10.3.10.83:5090 \
    -e BAMBOO_ZK_HOST=10.3.10.83:5092 \
    -e BAMBOO_ZK_PATH=/bb_gateway \
    -e BIND=":5090" \
    -e CONFIG_PATH="config/production.json" \
    --name=omega-bamboo \
    registry.shurenyun.com/bamboo-0.2.14:omega.v2.3
