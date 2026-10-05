FROM python:3.12-alpine
RUN apk add --no-cache bash curl docker-cli docker-cli-compose
WORKDIR /workspace
CMD ["bash", "scripts/acceptance.sh"]
