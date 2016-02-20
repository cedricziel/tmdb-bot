FROM alpine:3.3

RUN apk add --update bash curl && rm -rf /var/cache/apk/*
ADD entrypoint.sh /
ADD tmdb-bot /

WORKDIR /

CMD ["/entrypoint.sh"]
