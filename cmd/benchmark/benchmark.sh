go build ./main.go

# each test is ran with the first argument being the name , the second whether the endpoint accepts metrics, the third for the duration and the last being the discovery
# endpont. See test.river for details on each endpoint.
./main churn true 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.churn/discovery"
./main churn false 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.churn/discovery"

./main single true 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.single/discovery" 
./main single false 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.single/discovery"

./main many true 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.many/discovery" 
./main many false 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.many/discovery"

./main large true 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.large/discovery" 
./main large false 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.large/discovery"

