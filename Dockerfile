# APP_FILES picks where the panel and its binaries come from: app-source compiles them,
# app-release repackages the release tarballs docker.yml unpacks into docker-dist/.
ARG APP_FILES=app-source

# ========================================================
# Stage: Frontend (Vite)
# ========================================================
FROM --platform=$BUILDPLATFORM node:26-alpine AS frontend
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
COPY internal/web/translation /src/internal/web/translation
RUN npm run build

# ========================================================
# Stage: Builder (compiles under emulation on non-native platforms)
# ========================================================
FROM golang:1.27-alpine AS app-source
WORKDIR /app
ARG TARGETARCH
ARG TARGETVARIANT

RUN apk --no-cache --update add \
  build-base \
  gcc \
  curl \
  git \
  unzip

COPY . .
COPY --from=frontend /src/internal/web/dist ./internal/web/dist

ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"
RUN go build -ldflags "-w -s" -o build/x-ui main.go
RUN ./DockerInit.sh "$TARGETARCH" "$TARGETVARIANT"

# ========================================================
# Stage: Release binaries (runs natively, no compiling)
# ========================================================
FROM --platform=$BUILDPLATFORM alpine AS app-release
ARG TARGETARCH
ARG TARGETVARIANT
COPY docker-dist/ /dist/
# Release tarballs use install.sh's arch names; the panel wants xray-linux-arm32 and mtg-linux-arm.
RUN set -e; \
  case "$TARGETARCH/${TARGETVARIANT:-}" in \
    amd64/*) arch=amd64 ;; \
    arm64/*) arch=arm64 ;; \
    arm/v7 | arm/) arch=armv7 ;; \
    arm/v6) arch=armv6 ;; \
    386/*) arch=386 ;; \
    *) echo "no release tarball for $TARGETARCH/${TARGETVARIANT:-}" >&2; exit 1 ;; \
  esac; \
  tar -xzf "/dist/x-ui-linux-$arch.tar.gz" -C /tmp; \
  mkdir -p /app/build; \
  mv /tmp/x-ui/x-ui /tmp/x-ui/bin /app/build/; \
  case "$arch" in \
    armv*) \
      mv "/app/build/bin/xray-linux-$arch" /app/build/bin/xray-linux-arm32; \
      mv "/app/build/bin/mtg-linux-$arch" /app/build/bin/mtg-linux-arm ;; \
  esac

FROM ${APP_FILES} AS app-files

# ========================================================
# Stage: Final Image of 3x-ui
# ========================================================
FROM alpine
ENV TZ=Asia/Tehran
WORKDIR /app

RUN apk add --no-cache --update \
  ca-certificates \
  tzdata \
  fail2ban \
  bash \
  curl \
  openssl

COPY --from=app-files /app/build/ /app/
COPY DockerEntrypoint.sh /app/
COPY x-ui.sh /usr/bin/x-ui
COPY internal/web/translation /app/internal/web/translation


# Configure fail2ban
RUN rm -f /etc/fail2ban/jail.d/alpine-ssh.conf \
  && cp /etc/fail2ban/jail.conf /etc/fail2ban/jail.local \
  && sed -i "s/^\[ssh\]$/&\nenabled = false/" /etc/fail2ban/jail.local \
  && sed -i "s/^\[sshd\]$/&\nenabled = false/" /etc/fail2ban/jail.local \
  && sed -i "s/#allowipv6 = auto/allowipv6 = auto/g" /etc/fail2ban/fail2ban.conf

RUN chmod +x \
  /app/DockerEntrypoint.sh \
  /app/x-ui \
  /usr/bin/x-ui

ENV XUI_IN_DOCKER="true"
ENV XUI_MAIN_FOLDER="/app"
ENV XUI_ENABLE_FAIL2BAN="true"
ENV XUI_DB_TYPE=""
ENV XUI_DB_DSN=""
EXPOSE 2053
VOLUME [ "/etc/x-ui" ]
CMD [ "./x-ui" ]
ENTRYPOINT [ "/app/DockerEntrypoint.sh" ]
