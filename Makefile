.PHONY: help build docker-build docker-push docker-all clean test run

# Variables
APP_NAME := orgmserver
IMAGE_REGISTRY := orgmcr.or-gm.com
IMAGE_NAMESPACE := osmargm1202
IMAGE_NAME := $(IMAGE_REGISTRY)/$(IMAGE_NAMESPACE)/$(APP_NAME)
IMAGE_TAG := latest
FULL_IMAGE := $(IMAGE_NAME):$(IMAGE_TAG)

# Colores para output
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m # No Color

help: ## Muestra esta ayuda
	@echo "$(GREEN)ORGMServer - Makefile$(NC)"
	@echo ""
	@echo "Uso: make [objetivo]"
	@echo ""
	@echo "Objetivos disponibles:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}'

build: ## Compila el binario de Go
	@echo "$(GREEN)Compilando $(APP_NAME)...$(NC)"
	@go build -o $(APP_NAME) .
	@echo "$(GREEN)✓ Build completado: ./$(APP_NAME)$(NC)"

test: ## Ejecuta los tests
	@echo "$(GREEN)Ejecutando tests...$(NC)"
	@go test -v ./...

clean: ## Limpia archivos generados
	@echo "$(GREEN)Limpiando archivos...$(NC)"
	@rm -f $(APP_NAME)
	@rm -rf logs/*.log
	@echo "$(GREEN)✓ Limpieza completada$(NC)"

run: ## Ejecuta el servicio localmente (requiere variables de entorno)
	@echo "$(GREEN)Ejecutando $(APP_NAME)...$(NC)"
	@go run main.go --debug

docker-build: ## Construye la imagen Docker
	@echo "$(GREEN)Construyendo imagen Docker: $(FULL_IMAGE)...$(NC)"
	@docker build -t $(FULL_IMAGE) .
	@echo "$(GREEN)✓ Imagen construida: $(FULL_IMAGE)$(NC)"

docker-push: ## Sube la imagen al registro
	@echo "$(GREEN)Subiendo imagen al registro: $(FULL_IMAGE)...$(NC)"
	@docker push $(FULL_IMAGE)
	@echo "$(GREEN)✓ Imagen subida exitosamente$(NC)"

docker-all: docker-build docker-push ## Construye y sube la imagen Docker

docker-login: ## Inicia sesión en el registro Docker
	@echo "$(GREEN)Iniciando sesión en $(IMAGE_REGISTRY)...$(NC)"
	@docker login $(IMAGE_REGISTRY)
	@echo "$(GREEN)✓ Sesión iniciada$(NC)"

docker-run: ## Ejecuta el contenedor localmente (requiere variables de entorno)
	@echo "$(GREEN)Ejecutando contenedor...$(NC)"
	@docker run --rm -it \
		--env-file .env \
		-v $(PWD)/state:/tmp \
		-v $(PWD)/logs:/root/logs \
		$(FULL_IMAGE) --debug

docker-stop: ## Detiene y elimina el contenedor
	@echo "$(GREEN)Deteniendo contenedor...$(NC)"
	@docker stop $(APP_NAME) 2>/dev/null || true
	@docker rm $(APP_NAME) 2>/dev/null || true
	@echo "$(GREEN)✓ Contenedor detenido$(NC)"

deps: ## Descarga las dependencias de Go
	@echo "$(GREEN)Descargando dependencias...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✓ Dependencias actualizadas$(NC)"

fmt: ## Formatea el código
	@echo "$(GREEN)Formateando código...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✓ Código formateado$(NC)"

vet: ## Ejecuta go vet para verificar el código
	@echo "$(GREEN)Verificando código con go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✓ Verificación completada$(NC)"

lint: vet ## Ejecuta verificaciones de código (vet)

all: clean deps build test ## Limpia, descarga deps, compila y ejecuta tests

