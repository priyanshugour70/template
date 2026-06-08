FROM node:22-alpine

WORKDIR /app

RUN npm install -g pnpm@10.32.1 pm2

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .
RUN pnpm run build

ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

EXPOSE 3000

CMD ["pm2-runtime", "start", "ecosystem.config.cjs"]
