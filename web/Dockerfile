FROM node:24-alpine AS build
ARG BACKEND_INTERNAL_URL=http://127.0.0.1:8080
ENV BACKEND_INTERNAL_URL=$BACKEND_INTERNAL_URL
RUN corepack enable
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY . .
RUN pnpm build

FROM node:24-alpine
RUN corepack enable
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /app ./
EXPOSE 3000
CMD ["pnpm", "start"]
