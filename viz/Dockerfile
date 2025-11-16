FROM node:22-alpine AS build
WORKDIR /app

ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

COPY package.json package-lock.json* ./
COPY . .

RUN npm install --legacy-peer-deps
RUN npm run build

FROM node:22-alpine
WORKDIR /app
COPY --from=build /app /app
EXPOSE 7777

ENV NODE_ENV=production

CMD ["node", "build"]