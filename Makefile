redis=redis://localhost:6379

broker=$(redis)
lock=$(redis)
default_queue=$(redis)
result_backend=$(redis)

run:
	export BROKER=$(broker); \
	export LOCK=$(lock); \
	export DEFAULT_QUEUE=$(default_queue); \
	export RESULT_BACKEND=$(result_backend); \
	$(GOPATH)/bin/k8s-zero-scaler
