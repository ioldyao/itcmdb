# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# 设置npm镜像源为阿里云
RUN npm config set registry https://mirrors.aliyun.com/npm/

COPY package*.json ./
RUN npm install

COPY . .
RUN npm run build

# Production stage
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY deploy/nginx/default.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
