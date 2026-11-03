# ---- stage 1: build ----
    FROM golang:1.27.3 AS build

    WORKDIR /src
    
    COPY go.mod go.sum ./
    
    RUN go mod download
    
    COPY . .
    
    ARG VERSION=dev
    
    RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X github.com/phanivarun05/yamlchecker/cmd.version=${VERSION}" -o /out/yamlchecker .
    
    # ---- stage 2: runtime ----
    FROM gcr.io/distroless/static-debian13:nonroot
    
    # 1. Fixed the compilation extraction source path to match your go build output (/out/yamlchecker)
    # 2. Used a static absolute path (/yamlchecker) inside the final runtime layer
    COPY --from=build /out/yamlchecker /yamlchecker
    
    EXPOSE 2112
    
    # 2. Updated to an absolute execution route to eliminate any working directory reliance
    ENTRYPOINT [ "/yamlchecker" ]

    LABEL org.opencontainers.image.source=https://github.com/phanivarun05/yamlchecker
    LABEL org.opencontainers.image.description="YAML Checker is a tool that checks the validity of YAML files."
    LABEL org.opencontainers.image.licenses=MIT
    LABEL org.opencontainers.image.authors="Phani Varun <phanivarun05@gmail.com>"
    LABEL org.opencontainers.image.vendor="Phani Varun"
    LABEL org.opencontainers.image.version=${VERSION}
    LABEL org.opencontainers.image.revision=${COMMIT_SHA}
    LABEL org.opencontainers.image.created=${BUILD_DATE}
    LABEL org.opencontainers.image.url=https://github.com/phanivarun05/yamlchecker
    LABEL org.opencontainers.image.documentation=https://github.com/phanivarun05/yamlchecker/blob/main/README.md
    