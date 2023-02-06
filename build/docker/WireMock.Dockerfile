FROM wiremock/wiremock

COPY /build/docker/wiremockMappings/* /home/wiremock/mappings/
COPY /build/docker/wiremockFiles/* /home/wiremock/__files/