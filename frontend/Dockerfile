FROM node:20-alpine AS build-stage

# Set the working directory inside the container
WORKDIR /app

# Copy package.json and package-lock.json to install dependencies
COPY package*.json ./

# Install dependencies
RUN npm install
# Copy the rest of the application code
COPY . .

# Run the build command
RUN npm run build

FROM caddy:alpine AS production-build

# Copy the built Vue.js assets into Caddy's default serving directory
COPY --from=build-stage /app/dist /usr/share/caddy

# Add a Caddyfile to handle SPA routing
COPY Caddyfile /etc/caddy/Caddyfile

EXPOSE 80
EXPOSE 443

ENTRYPOINT ["caddy", "run", "--config", "/etc/caddy/Caddyfile"]