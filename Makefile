.PHONY:

up:
	docker compose up --build -d

down:
	docker compose down

down-v:
	docker compose down -v

build-scheduler:
	docker build --target scheduler -t distributed-scheduler-scheduler:latest .

restart-scheduler:
	kubectl -n distributed-scheduler delete deployments.apps scheduler
	kubectl apply -f k8s/apps/scheduler.yaml

get-pods:
	kubectl -n distributed-scheduler get pod

build-worker:
	docker build --target worker -t distributed-scheduler-worker:latest .

restart-worker:
	kubectl -n distributed-scheduler delete deployments.apps worker
	kubectl apply -f k8s/apps/worker.yaml