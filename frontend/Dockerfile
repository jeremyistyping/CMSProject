FROM bitnami/node:22.12.0 as builder 

COPY . /app 
COPY ./.env.example /app/.env
WORKDIR /app 
RUN npm install 


RUN npm run build
EXPOSE 3000
CMD ["npm","run","start"]
