FROM node:22-alpine AS build
WORKDIR /app

ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

# Build-time copy of config from repository root (make sure build context
# includes the repo root so this file is available during build).
COPY imagine.json /app/imagine.json
ENV IMAGINE_CONFIG_PATH=/app/imagine.json

# Copy package.json and lock from the viz subdir because build context is the repo root
COPY viz/package.json viz/package-lock.json* ./
COPY viz/ ./

RUN npm install --legacy-peer-deps
RUN npm run build

FROM node:22-alpine
WORKDIR /app
COPY --from=build /app /app
EXPOSE 7777

ENV NODE_ENV=production

CMD ["node", "build"]