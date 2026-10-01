FROM node:22-alpine AS deps
RUN corepack enable
WORKDIR /repo
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/web/package.json apps/web/
COPY packages/ui/package.json packages/ui/
COPY packages/types/package.json packages/types/
COPY packages/config/package.json packages/config/
RUN pnpm install --frozen-lockfile

FROM deps AS build
COPY . .
ARG API_URL=http://api:8080
ENV NEXT_TELEMETRY_DISABLED=1 API_URL=$API_URL
RUN pnpm --filter @listou/web build

FROM node:22-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production NEXT_TELEMETRY_DISABLED=1 PORT=3000 HOSTNAME=0.0.0.0
COPY --from=build /repo/apps/web/.next/standalone ./
COPY --from=build /repo/apps/web/.next/static ./apps/web/.next/static
COPY --from=build /repo/apps/web/public ./apps/web/public
USER node
EXPOSE 3000
CMD ["node", "apps/web/server.js"]
