FROM public.ecr.aws/docker/library/golang:1.23 AS build

WORKDIR /go/src/actioneer-operator
COPY go.mod .

RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 go build -o /go/bin/actioneer-operator main.go

FROM gcr.io/distroless/static-debian12:latest
COPY --from=build /go/bin/actioneer-operator /actioneer-operator

# expose http server
EXPOSE 8001
# internal stuff such as metrics
EXPOSE 2112

CMD ["./actioneer-operator"]
