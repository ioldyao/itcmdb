# Build stage
FROM node:20-alpine AS builder

# 设置 Alpine 镜像源为阿里云
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

WORKDIR /app

# 设置 npm 镜像源为淘宝镜像
COPY package*.json ./
RUN npm config set registry https://registry.npmmirror.com && npm install -d
# RUN npm install 

COPY . .
RUN npm run build

# Production stage
FROM nginx:alpine

# 设置 Alpine 镜像源为阿里云
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

COPY --from=builder /app/dist /usr/share/nginx/html
COPY deploy/nginx/default.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
