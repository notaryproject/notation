FROM notaryv2/registry:nv2-prototype-2

RUN apk add openssl

COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh
ENTRYPOINT [ "/entrypoint.sh" ]
CMD [ "serve", "/etc/docker/registry/config.yml" ]
