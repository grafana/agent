docker stop ava
docker rm ava
go build ./main.go
docker run --name=ava -d -p 9001:9001 quay.io/freshtracks.io/avalanche --metric-count=20000
./main 20000 true 1h  
./main 20000 false 1h
docker stop ava
docker rm ava

# docker run --name=ava -d -p 9001:9001 quay.io/freshtracks.io/avalanche --metric-count=20000
#./main 20000 true 1h
#./main 20000 false 1h
#docker stop ava
#docker rm ava

# docker run --name=ava -d -p 9001:9001 quay.io/freshtracks.io/avalanche --metric-count=25000
# ./main 25000 true 1h
# ./main 25000 false 1h
docker stop ava
docker rm ava
