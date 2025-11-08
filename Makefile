.PHONY:

up:
	docker compose up --build -d

down:
	docker compose down

down-v:
	docker compose down -v