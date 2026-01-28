FROM python:3.11-alpine

WORKDIR /app

# 设置 Alpine 镜像源为阿里云
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# Install build dependencies for pydantic-core
RUN apk add --no-cache \
    gcc \
    musl-dev \
    libffi-dev \
    rust \
    cargo \
    ca-certificates

# Copy requirements
COPY requirements.txt .

# Install dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy workflow engine code
COPY workflow_engine /usr/local/lib/python3.11/site-packages/workflow_engine

# Set environment
ENV PYTHONPATH=/usr/local/lib/python3.11/site-packages
ENV WORKFLOW_ENGINE_HOST=0.0.0.0
ENV WORKFLOW_ENGINE_PORT=8000

EXPOSE 8000

CMD ["python3", "-m", "workflow_engine.api"]
