.PHONY: prod-build prod-up prod-down prod-restart prod-logs prod-status prod-clean

POD          := reconta-prod
WEB_PORT     := 8080
API_IMAGE    := reconta-api:local
WEB_IMAGE    := reconta-web:local
DATA_VOLUME  := reconta-api-data
ENV_FILE     := $(if $(wildcard api/.env),api/.env,api/.env.example)

## Builda as imagens da API (Go) e do Web (Vue via nginx)
prod-build:
	podman build -t $(API_IMAGE) -f deploy/podman/Containerfile.api .
	podman build -t $(WEB_IMAGE) -f deploy/podman/Containerfile.web .

## Sobe API + Nginx/Web no mesmo pod, simulando o ambiente de produção
prod-up: prod-build
	podman pod exists $(POD) && podman pod rm -f $(POD) || true
	podman pod create --name $(POD) -p $(WEB_PORT):$(WEB_PORT)
	podman volume exists $(DATA_VOLUME) || podman volume create $(DATA_VOLUME)
	# ENV=development apenas para desativar cookies Secure (não há HTTPS local);
	# build, binário e nginx continuam idênticos aos de produção.
	podman run -d --pod $(POD) --name reconta-prod-api \
		--env-file $(ENV_FILE) \
		-e PORT=3020 \
		-e ENV=development \
		-e APP_URL=http://localhost:$(WEB_PORT) \
		-e DB_PATH=./data/reconta.db \
		-v $(DATA_VOLUME):/app/data \
		$(API_IMAGE)
	podman run -d --pod $(POD) --name reconta-prod-web $(WEB_IMAGE)
	@echo "✅ Disponível em http://localhost:$(WEB_PORT)"

## Para e remove o pod (mantém o volume com o banco de dados)
prod-down:
	podman pod rm -f $(POD) || true

## Reinicia containers já buildados, sem rebuild
prod-restart:
	podman pod restart $(POD)

## Segue os logs de API e Web
prod-logs:
	podman pod logs -f $(POD)

## Mostra o status do pod e containers
prod-status:
	podman pod ps --filter name=$(POD)
	podman ps --pod --filter pod=$(POD)

## Remove pod, imagens e volume (apaga o banco de dados local)
prod-clean: prod-down
	podman rmi -f $(API_IMAGE) $(WEB_IMAGE) || true
	podman volume rm -f $(DATA_VOLUME) || true
