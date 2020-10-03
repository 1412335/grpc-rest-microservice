FROM node:13-alpine

RUN apk add --no-cache python

WORKDIR /app

COPY package*.json .

RUN npm install

COPY . .

# RUN npx webpack /app/cmd/client.js
# RUN npx webpack --config webpack.config.js

EXPOSE 8081
# CMD ["python", "-m", "SimpleHTTPServer", "8081"]

# CMD npx webpack-dev-server --config webpack.config.js
CMD npm start