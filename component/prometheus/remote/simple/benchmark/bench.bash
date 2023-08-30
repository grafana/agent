docker stop ava
docker rm ava
go build ./main.go
docker run --name=ava -d -p 9001:9001 quay.io/freshtracks.io/avalanche --metric-count=1000
./main 1000 true 1h
./main 1000 false 1h
docker stop ava
docker rm ava

docker run --name=ava -d -p 9001:9001 quay.io/freshtracks.io/avalanche --metric-count=10000
./main 10000 true 1h
./main 10000 false 1h
docker stop ava
docker rm ava

docker run --name=ava -d -p 9001:9001 quay.io/freshtracks.io/avalanche --metric-count=25000
./main 25000 true 1h
./main 25000 false 1h
docker stop ava
docker rm ava