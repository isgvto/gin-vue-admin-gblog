# syntax=docker/dockerfile:1
FROM node:24-alpine AS build

WORKDIR /app

COPY package*.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci --prefer-offline --no-audit --no-fund \
    --fetch-retries=5 \
    --fetch-retry-mintimeout=20000 \
    --fetch-retry-maxtimeout=120000 \
    --fetch-timeout=300000

COPY . .
RUN npm run build


FROM nginx:alpine

ARG API_PROXY_PASS=http://gva-gbserver:8888

COPY nginx.conf.template /etc/nginx/templates/default.conf.template
COPY --from=build /app/dist /usr/share/nginx/html

ENV API_PROXY_PASS=${API_PROXY_PASS}
