FROM golang:1.10 as build

ADD . /go/src/github.com/google/certificate-transparency-go
WORKDIR /go/src/github.com/google/certificate-transparency-go

ARG GOFLAGS=""
RUN go get ${GOFLAGS} ./trillian/ctfe/ct_server

FROM gcr.io/distroless/base

COPY --from=build /go/bin/ct_server /
# Include the source code for the CTFE and all of its transitive dependencies.
# This ensures that we meet their license requirements to redistribute their
# licenses, copyright notices and, in some cases, source code.
COPY --from=build /go/src /go/src


ENTRYPOINT ["/ct_server"]
