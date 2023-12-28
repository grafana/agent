go build ./main.go


./main churn true 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.churn/discovery"
./main churn false 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.churn/discovery"

./main single true 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.single/discovery" 
./main single false 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.single/discovery"

./main many true 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.many/discovery" 
./main many false 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.many/discovery"

./main large true 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.large/discovery" 
./main large false 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.large/discovery"

# docker run --name=ava -d -p 9001:9001 quay.io/freshtracks.io/avalanche --metric-count=20000
#./main 20000 true 1h
#./main 20000 false 1h

# docker run --name=ava -d -p 9001:9001 quay.io/freshtracks.io/avalanche --metric-count=25000
# ./main 25000 true 1h
# ./main 25000 false 1h
