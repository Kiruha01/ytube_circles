#!/bin/bash

# Проверка флага -y
AUTO_YES=false
if [ "$1" = "-y" ]; then
    AUTO_YES=true
    shift  # Сдвигаем аргументы, чтобы версия стала $1
fi

# Проверка, что аргумент с версией передан
if [ -z "$1" ]; then
    echo "Ошибка: укажите версию как аргумент (например, ./build_docker.sh [-y] 1.0)"
    exit 1
fi

# Установка переменных
PROJECT_NAME="yt_circles"
DOCKER_USERNAME="kiruha01"
VERSION="$1"  # Версия берётся из первого аргумента (или второго, если есть -y)

# Полный тег для нового билда
NEW_TAG="${DOCKER_USERNAME}/${PROJECT_NAME}:${VERSION}"
LATEST_TAG="${DOCKER_USERNAME}/${PROJECT_NAME}:latest"

# Сборка Docker-образа с новым тегом
echo "Собираем образ с тегом ${NEW_TAG}..."
docker build -t "${NEW_TAG}" .

# Проверка успешности сборки
if [ $? -eq 0 ]; then
    echo "Сборка успешно завершена"
else
    echo "Ошибка при сборке образа"
    exit 1
fi

# Функция для получения ответа
get_answer() {
    local prompt="$1"
    if [ "$AUTO_YES" = true ]; then
        echo "$prompt Y (автоматический ответ из-за флага -y)"
        echo "y"
    else
        read -p "$prompt" answer
        echo "${answer:-y}"
    fi
}

# Запрос на обновление тега latest
UPDATE_LATEST=$(get_answer "Обновить тег latest? (Y/n): ")
if [ "$UPDATE_LATEST" = "n" ] || [ "$UPDATE_LATEST" = "N" ]; then
    echo "Тег latest не будет обновлён"
else
    echo "Обновляем тег latest..."
    docker tag "${NEW_TAG}" "${LATEST_TAG}"
fi

# Запрос на пуш нового тега
PUSH_NEW=$(get_answer "Отправить новый тег ${NEW_TAG} в реестр? (Y/n): ")
if [ "$PUSH_NEW" = "n" ] || [ "$PUSH_NEW" = "N" ]; then
    echo "Новый тег ${NEW_TAG} не будет отправлен"
else
    echo "Отправляем новый тег ${NEW_TAG} в реестр..."
    if ! docker push "${NEW_TAG}"; then
      echo "Ошибка при отправке тега ${NEW_TAG}"
      exit 1
    fi
fi

# Если latest обновлён, запрос на его пуш
if [ "$UPDATE_LATEST" = "y" ] || [ "$UPDATE_LATEST" = "Y" ]; then
    PUSH_LATEST=$(get_answer "Отправить тег latest в реестр? (Y/n): ")
    if [ "$PUSH_LATEST" = "y" ] || [ "$PUSH_LATEST" = "Y" ]; then
        echo "Отправляем тег latest в реестр..."
        if ! docker push "${LATEST_TAG}"; then
          echo "Ошибка при отправке тега ${NEW_TAG}"
          exit 1
        fi
    else
        echo "Тег latest не будет отправлен"
    fi
fi

echo "Готово! Новый образ: ${NEW_TAG}"
if [ "$UPDATE_LATEST" = "y" ] || [ "$UPDATE_LATEST" = "Y" ]; then
    echo "Тег latest обновлён: ${LATEST_TAG}"
fi
