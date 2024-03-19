redis-server --cluster-enabled yes --cluster-config-file nodes.conf --cluster-node-timeout 5000 --requirepass 29102087 &
while [[ $(redis-cli --pass 29102087 ping) != 'PONG' ]]
do
  echo "Node $node not ready, waiting for all the nodes to be ready..."
  sleep 1
done  
redis-cli --pass 29102087 --cluster create 127.0.0.1:6379 127.0.0.1:6379 127.0.0.1:6379 --cluster-yes 
tail -f /dev/null






