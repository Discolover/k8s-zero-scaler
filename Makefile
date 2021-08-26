push:
	docker build -t semars/k8s-zero-scaler:6.0.0 .
	docker push semars/k8s-zero-scaler:6.0.0

deploy:
	kubectl apply -f zs-pod.yml
	sleep 10
	kubectl apply -f samples/add_pod_creation_notify_initcontainer.yml
	sleep 5
	kubectl apply -f samples/sample_deployment.yml

clear:
	kubectl delete -f zs-pod.yml
	kubectl delete -f samples/add_pod_creation_notify_initcontainer.yml
	kubectl delete -f samples/sample_deployment.yml