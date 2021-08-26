push:
	docker build -t semars/k8s-zero-scaler .
	docker push semars/k8s-zero-scaler:latest
deploy:
	kubectl apply -f zs-pod.yml