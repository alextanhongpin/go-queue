run:
	@open http://localhost:8080
	@go run main.go

up:
	@docker-compose up -d

down:
	@docker-compose down
