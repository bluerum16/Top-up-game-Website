# ════════════════════════════════════════════
# Stage 1: Dependencies
# Install node_modules (di-cache jika package.json tidak berubah)
# ════════════════════════════════════════════
FROM node:20-alpine AS deps

RUN apk add --no-cache libc6-compat

WORKDIR /app

# Copy package files dulu → layer cache optimal
COPY package.json package-lock.json ./

# npm ci = install bersih dari lock file (lebih deterministik dari npm install)
RUN npm ci

# ════════════════════════════════════════════
# Stage 2: Builder
# Build Next.js production bundle
# ════════════════════════════════════════════
FROM node:20-alpine AS builder

WORKDIR /app

# Ambil node_modules dari stage deps
COPY --from=deps /app/node_modules ./node_modules

# Copy seluruh source
COPY . .

# Nonaktifkan telemetry Next.js
ENV NEXT_TELEMETRY_DISABLED=1

# Build → output: .next/standalone (diset di next.config.js)
RUN npm run build

# ════════════════════════════════════════════
# Stage 3: Runner — image produksi final
# ════════════════════════════════════════════
FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

# User non-root
RUN addgroup --system --gid 1001 nodejs && \
    adduser  --system --uid 1001 nextjs

# Salin output standalone Next.js
# Catatan: next.config.js harus punya: output: 'standalone'
COPY --from=builder /app/public                              ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone  ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static      ./.next/static

USER nextjs

EXPOSE 3000
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD wget -qO- http://localhost:3000/api/health || exit 1

CMD ["node", "server.js"]