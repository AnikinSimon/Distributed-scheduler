FROM nginx:alpine

# Копируем конфигурацию
COPY nginx.conf /etc/nginx/nginx.conf

#COPY ssl/ /etc/nginx/ssl/

# Создаем директорию для приложений
RUN mkdir -p /var/www/html

EXPOSE 80
EXPOSE 443

CMD ["nginx", "-g", "daemon off;"]