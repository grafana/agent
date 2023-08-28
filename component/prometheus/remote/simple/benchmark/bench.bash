docker stop ava
docker rm ava
go build ./main.go
docker run --name=ava -d -p 9001:9001 quay.io/freshtracks.io/avalanche --metric-count=100
./main 1000 true 1h
./main 1000 false 1h
docker stop ava
docker rm ava

docker run --name=ava -d -p 9001:9001 quay.io/freshtracks.io/avalanche --metric-count=100
./main 10000 true 1h
./main 10000 false 1h
docker stop ava
docker rm ava

docker run --name=ava -d -p 9001:9001 quay.io/freshtracks.io/avalanche --metric-count=100
./main 100000 true 1h
./main 100000 false 1h
docker stop ava
docker rm ava
