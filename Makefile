.PHONY: help build up down logs deploy init-db

help:
	@echo "SteelLink Admin - Deployment Commands"
	@echo ""
	@echo "Usage:"
	@echo "  make build      Build Docker images"
	@echo "  make up         Start services"
	@echo "  make down       Stop services"
	@echo "  make logs       View logs"
	@echo "  make deploy     Build and start services"
	@echo "  make init-db    Initialize database (run init-db.sql)"
	@echo ""

build:
	docker compose -f docker-compose.prod.yml build

up:
	docker compose -f docker-compose.prod.yml up -d

down:
	docker compose -f docker-compose.prod.yml down

logs:
	docker compose -f docker-compose.prod.yml logs -f

deploy: build up
	@echo "Deployment complete. Access the admin panel at http://localhost"

init-db:
	@echo "Initializing database..."
	@read -p "Enter MySQL password: " pwd; \
	mysql -h $(DB_HOST) -P $(DB_PORT) -u $(DB_USER) -p$$pwd $(DB_NAME) < scripts/init-db.sql

# Development commands
dev-frontend:
	cd frontend && pnpm dev

dev-backend:
	cd backend && go run cmd/server/main.go
