# Build a thin MySQL image that automatically loads schema.sql on first init
FROM mysql:8.0

# Recommended defaults; safe for dev
ENV MYSQL_ROOT_PASSWORD=root \
    MYSQL_DATABASE=sensors \
    MYSQL_USER=app \
    MYSQL_PASSWORD=app

# Place schema where MySQL entrypoint will auto-run it on first init
# (only when the data dir is empty)
COPY schema.sql /docker-entrypoint-initdb.d/01_schema.sql