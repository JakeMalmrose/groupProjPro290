# syntax=docker/dockerfile:1

ARG NODE_VERSION=18.16.1

FROM node:${NODE_VERSION}-alpine

# Use production node environment by default.
ENV NODE_ENV production

WORKDIR /usr/src/app

# Copy package.json and package-lock.json (if available) to the working directory
COPY package*.json ./

# Install dependencies based on the presence of package-lock.json
RUN --mount=type=cache,target=/root/.npm \
    if [ -f package-lock.json ]; then \
      npm ci --omit=dev; \
    else \
      npm install --omit=dev; \
    fi

# Copy the necessary files to the working directory
COPY . .

# Expose the port on which the service will run
EXPOSE 8080

# Define the command to run the service
CMD npm start