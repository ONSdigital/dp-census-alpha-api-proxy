FROM golang:1.14

ARG BIND_ADDR
ARG AUTH_TOKEN
ARG FTB_URL

# Create app directory
WORKDIR /app

ENV HUMAN_LOG 1
ENV BIND_ADDR :${BIND_ADDR}
ENV AUTH_TOKEN ${AUTH_TOKEN}
ENV FTB_URL ${FTB_URL}

COPY build/dp-census-alpha-api-proxy /app/

CMD ./dp-census-alpha-api-proxy