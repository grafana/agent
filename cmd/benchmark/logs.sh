go build -o main  

# each test is ran with the first argument being the name , the second whether the endpoint accepts metrics, the third for the duration and the last being the discovery
# endpont. See test.river for details on each endpoint.
./main logs 1h
