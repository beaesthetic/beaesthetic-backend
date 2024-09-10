FROM node:20-alpine
WORKDIR /app

COPY . .
RUN npm install

ENTRYPOINT ["node", "/app/agenda.js"]