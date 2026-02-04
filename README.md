# go-docker-demo
go docker demo

kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=zhang12590 \
  --docker-password=<your-github-token> \
  --docker-email=zhangyouyu12590@163.com \
  --namespace=default
