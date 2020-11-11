FROM golang:alpine AS build
RUN apk add make
WORKDIR /src
COPY ./ ./
RUN make clean && make alpine-build

FROM alpine AS bin-unix
# RUN apk add -qU openssh
# RUN mkdir -p /root/.ssh/ && \
#     ssh-keyscan github.com >> ~/.ssh/known_hosts
COPY --from=build /src/build/vercheck /vercheck
ENTRYPOINT ["/vercheck"]
