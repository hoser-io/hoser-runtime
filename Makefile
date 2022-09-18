install:
	go install ./...

docker:
	docker build -t hoser:latest .

docker-al2:
	docker build --platform=linux/amd64 -t hoser.io/hoser:latest-amd64 .

publish: docker
	docker tag $$(docker images -q hoser:latest-amd64) 792169137994.dkr.ecr.us-east-2.amazonaws.com/hoser-io/basic
	aws ecr get-login-password --region us-east-2 --profile hoserio | docker login --username AWS --password-stdin 792169137994.dkr.ecr.us-east-2.amazonaws.com
	docker push 792169137994.dkr.ecr.us-east-2.amazonaws.com/hoser-io/basic