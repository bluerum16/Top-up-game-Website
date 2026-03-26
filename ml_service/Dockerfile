# ════════════════════════════════════════════
# Stage 1: Builder — install dependencies
# ════════════════════════════════════════════
FROM python:3.11-slim AS builder

WORKDIR /app

# Install build tools (hanya dibutuhkan saat install deps, tidak perlu di runtime)
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements dulu → layer cache
# Jika requirements.txt tidak berubah, pip install di-skip
COPY requirements.txt .

# Install ke folder lokal (nanti di-copy ke stage runner)
RUN pip install --no-cache-dir --prefix=/install -r requirements.txt

# ════════════════════════════════════════════
# Stage 2: Runner — image produksi final
# ════════════════════════════════════════════
FROM python:3.11-slim AS runner

WORKDIR /app

# Install runtime libs saja (libpq dibutuhkan oleh psycopg2 saat runtime)
RUN apt-get update && apt-get install -y --no-install-recommends \
    libpq5 \
    && rm -rf /var/lib/apt/lists/*

# Copy installed packages dari builder stage
COPY --from=builder /install /usr/local

# User non-root
RUN addgroup --system --gid 1001 mlgroup && \
    adduser  --system --uid 1001 mluser --ingroup mlgroup

# Copy source code
COPY --chown=mluser:mlgroup . .

USER mluser

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=10s --start-period=20s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8000/health')" || exit 1

# reload=False di production, workers=2 sesuaikan dengan CPU VPS
CMD ["uvicorn", "main:app", \
     "--host", "0.0.0.0", \
     "--port", "8000", \
     "--workers", "2", \
     "--no-access-log"]