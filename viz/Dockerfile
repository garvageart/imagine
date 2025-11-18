# Builder stage
FROM node:latest AS deps
WORKDIR /app
ENV NODE_ENV=production

RUN corepack enable && corepack prepare pnpm@latest --activate
COPY package.json pnpm-lock.yaml .npmrc ./

RUN pnpm install --frozen-lockfile

FROM node:latest AS build
WORKDIR /app
ENV NODE_ENV=production

RUN corepack enable && corepack prepare pnpm@latest --activate

COPY --from=deps /app/node_modules ./node_modules
COPY . .

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

