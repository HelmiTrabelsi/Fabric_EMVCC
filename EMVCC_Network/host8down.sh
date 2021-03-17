docker-compose -f host8.yaml down -v
docker rm -f $(docker ps -a -q)
docker volume rm $(docker volume ls -q)
docker rmi $(docker images dev-* -q)
