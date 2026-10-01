FROM nginx:alpine

COPY . /usr/share/nginx/html

# Configuración de nginx para servir módulos ES6 correctamente
RUN echo 'server { \
    listen 80; \
    server_name localhost; \
    root /usr/share/nginx/html; \
    index index.html; \
    location / { \
        try_files $uri $uri/ /index.html; \
    } \
    location ~* \.(js|mjs)$ { \
        add_header Content-Type application/javascript; \
    } \
    location ~* \.css { \
        add_header Content-Type text/css; \
    } \
}' > /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
