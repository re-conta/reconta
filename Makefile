.PHONY: build up down restart logs status clean certs dev hosts-check

POD          := reconta
WEB_PORT     := 443
API_IMAGE    := reconta-api:local
WEB_IMAGE    := reconta-web:local
DATA_VOLUME  := reconta-api-data
ENV_FILE     := $(if $(wildcard api/.env),api/.env,api/.env.example)

DOMAIN       := reconta.local
CERT_DIR     := certs
CERT_FILE    := $(CERT_DIR)/$(DOMAIN).pem
KEY_FILE     := $(CERT_DIR)/$(DOMAIN)-key.pem

help: ## Mostra esta ajuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Builda as imagens da API (Go) e do Web (Vue via nginx)
	podman build -t $(API_IMAGE) -f podman/Containerfile.api .
	podman build -t $(WEB_IMAGE) -f podman/Containerfile.web .

up: build certs hosts-check ## Sobe API + Nginx/Web no mesmo pod (HTTPS em https://reconta.local), simulando produção
	podman pod exists $(POD) && podman pod rm -f $(POD) || true
	podman pod create --name $(POD) -p $(WEB_PORT):$(WEB_PORT)
	podman volume exists $(DATA_VOLUME) || podman volume create $(DATA_VOLUME)
	podman run -d --pod $(POD) --name reconta-api \
		--env-file $(ENV_FILE) \
		-e PORT=3020 \
		-e ENV=production \
		-e APP_URL=https://$(DOMAIN) \
		-e DB_PATH=./data/reconta.db \
		-v $(DATA_VOLUME):/app/data \
		$(API_IMAGE)
	podman run -d --pod $(POD) --name reconta-web \
		-v ./$(CERT_DIR):/etc/nginx/certs:ro \
		$(WEB_IMAGE)
	@echo "✅ Disponível em https://$(DOMAIN)"

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

certs: ## Gera certificado TLS autoassinado para reconta.local, se ainda não existir
	@if [ -f "$(CERT_FILE)" ] && [ -f "$(KEY_FILE)" ]; then \
		echo "✅ Certificado já existe em $(CERT_DIR)/"; \
	else \
		mkdir -p $(CERT_DIR); \
		openssl req -x509 -newkey rsa:2048 -nodes -days 825 \
			-keyout $(KEY_FILE) -out $(CERT_FILE) \
			-subj "/CN=$(DOMAIN)" \
			-addext "subjectAltName=DNS:$(DOMAIN),DNS:localhost,IP:127.0.0.1"; \
		echo "✅ Certificado gerado em $(CERT_DIR)/"; \
	fi

hosts-check: ## Verifica se reconta.local resolve para 127.0.0.1 em /etc/hosts
	@if getent hosts $(DOMAIN) | grep -q '^127\.0\.0\.1'; then \
		echo "✅ $(DOMAIN) resolve para 127.0.0.1"; \
	else \
		echo "⚠️  $(DOMAIN) não está em /etc/hosts. Adicione a linha abaixo:"; \
		echo "    127.0.0.1 $(DOMAIN)"; \
		echo "    sudo sh -c 'echo \"127.0.0.1 $(DOMAIN)\" >> /etc/hosts'"; \
	fi

dev: certs hosts-check ## Gera certificado local (se preciso) e sobe API + Vite em modo desenvolvimento
	@echo "✅ Disponível em https://$(DOMAIN):5173 (aceite o certificado autoassinado no navegador)"
	VITE_HTTPS=1 bun run dev
