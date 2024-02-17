# Dockerfile
FROM postgres:13

# Set environment variables
ENV POSTGRES_USER sqladmin
ENV POSTGRES_PASSWORD spersecretpassword
ENV POSTGRES_DB news

# Add initialization scripts
ADD init.sql /docker-entrypoint-initdb.d/