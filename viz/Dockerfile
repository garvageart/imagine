# Builder stage
FROM node:latest AS deps
WORKDIR /app
ENV NODE_ENV=production

RUN npm install -g pnpm@latest
COPY viz/package.json viz/pnpm-lock.yaml viz/.npmrc ./

RUN pnpm install --frozen-lockfile

FROM node:latest AS build
WORKDIR /app
ENV NODE_ENV=production
ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

RUN npm install -g pnpm@latest

COPY --from=deps /app/node_modules ./node_modules
COPY viz/ ./

RUN pnpm run build

# running state
FROM node:latest AS export
WORKDIR /app

# copying previously 
COPY --from=build /app/build /app/build
COPY --from=build /app/package.json /app/package.json
COPY --from=build /app/node_modules /app/node_modules
COPY --from=build /app/static /app/static

ENV PORT=7777
EXPOSE 7777

# running as non-root
USER node

CMD ["node", "/app/build"]

