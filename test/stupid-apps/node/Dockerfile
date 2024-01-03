FROM node:16-bullseye-slim

WORKDIR /usr/src/app

COPY ./test/stupid-apps/node/package*.json ./

RUN npm install

# Bundle app source
COPY ./test/stupid-apps/node/app2.js ./app2.js
COPY ./test/stupid-apps/node/subprocess1.js ./subprocess1.js
COPY ./test/stupid-apps/node/subprocess2.js ./subprocess2.js

CMD [ "node", "--perf-basic-prof", "app2.js" ]
