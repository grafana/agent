go build -o main

# each test is ran with the first argument being the name , the second whether the endpoint accepts metrics, the third for the duration and the last being the discovery
# endpont. See test.river for details on each endpoint.
./main metrics churn true 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.churn/discovery"
./main metrics churn false 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.churn/discovery"

./main metrics single true 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.single/discovery" 
./main metrics single false 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.single/discovery"

./main metrics many true 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.many/discovery" 
./main metrics many false 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.many/discovery"

./main metrics large true 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.large/discovery" 
./main metrics large false 1h "http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.large/discovery"

