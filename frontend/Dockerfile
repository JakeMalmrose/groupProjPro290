# syntax=docker/dockerfile:1

ARG NODE_VERSION=18.16.1

FROM node:${NODE_VERSION}-alpine

# Use production node environment by default.
ENV NODE_ENV production


WORKDIR /usr/src/app

RUN --mount=type=bind,source=package.json,target=package.json \
    --mount=type=bind,source=package-lock.json,target=package-lock.json \
    --mount=type=cache,target=/root/.npm \
    npm ci --omit=dev

# Copy the necessary files to the working directory
COPY . .

# Expose the port on which the service will run
EXPOSE 3000

# Define the command to run the service
CMD npm start