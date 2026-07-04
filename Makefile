.PHONY: build up down restart logs status clean

POD          := reconta
WEB_PORT     := 8080
API_IMAGE    := reconta-api:local
WEB_IMAGE    := reconta-web:local
DATA_VOLUME  := reconta-api-data
ENV_FILE     := $(if $(wildcard api/.env),api/.env,api/.env.example)

help: ## Mostra esta ajuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Builda as imagens da API (Go) e do Web (Vue via nginx)
	podman build -t $(API_IMAGE) -f deploy/podman/Containerfile.api .
	podman build -t $(WEB_IMAGE) -f deploy/podman/Containerfile.web .

up: build ## Sobe API + Nginx/Web no mesmo pod, simulando o ambiente de produção
	podman pod exists $(POD) && podman pod rm -f $(POD) || true
	podman pod create --name $(POD) -p $(WEB_PORT):$(WEB_PORT)
	podman volume exists $(DATA_VOLUME) || podman volume create $(DATA_VOLUME)
	# ENV=development apenas para desativar cookies Secure (não há HTTPS local);
	# build, binário e nginx continuam idênticos aos de produção.
	podman run -d --pod $(POD) --name reconta-api \
		--env-file $(ENV_FILE) \
		-e PORT=3020 \
		-e ENV=development \
		-e APP_URL=http://localhost:$(WEB_PORT) \
		-e DB_PATH=./data/reconta.db \
		-v $(DATA_VOLUME):/app/data \
		$(API_IMAGE)
	podman run -d --pod $(POD) --name reconta-web $(WEB_IMAGE)
	@echo "✅ Disponível em http://localhost:$(WEB_PORT)"

down: ## Para e remove o pod (mantém o volume com o banco de dados)
	podman pod rm -f $(POD) || true


restart: ## Reinicia containers já buildados, sem rebuild
	podman pod restart $(POD)

logs: ## Segue os logs de API e Web
	podman pod logs -f $(POD)

status: ## Mostra o status do pod e containers
	podman pod ps --filter name=$(POD)
	podman ps --pod --filter pod=$(POD)

clean: down ## Remove pod, imagens e volume (apaga o banco de dados local)
	podman rmi -f $(API_IMAGE) $(WEB_IMAGE) || true
	podman volume rm -f $(DATA_VOLUME) || true
