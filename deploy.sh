#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ -f "${SCRIPT_DIR}/deploy.env" ]]; then
  # shellcheck source=/dev/null
  source "${SCRIPT_DIR}/deploy.env"
fi

: "${SERVER:?Set SERVER in deploy.env or environment (see deploy.env.example)}"
: "${SSH_KEY:?Set SSH_KEY in deploy.env or environment (see deploy.env.example)}"

readonly REMOTE_APP_DIR="/opt/sumit"
readonly BINARY_NAME="sumit"
readonly PROD_ENV_KEYS=(GROQ_API_KEY)

write_env_prod() {
  local dest="${SCRIPT_DIR}/.env.prod"
  : > "${dest}"
  local key
  for key in "${PROD_ENV_KEYS[@]}"; do
    local val="${!key:-}"
    if [[ -n "${val}" ]]; then
      printf '%s=%s\n' "${key}" "${val}" >> "${dest}"
    fi
  done
}

echo "==> Generowanie .env.prod z deploy.env..."
write_env_prod

echo "==> Kompilacja dla Linux amd64..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o "${BINARY_NAME}" .

echo "==> Przesyłanie binarki na serwer..."
scp -i "${SSH_KEY}" "${BINARY_NAME}" "${SERVER}:/tmp/${BINARY_NAME}"

echo "==> Przesyłanie zasobów statycznych (static/, assets/)..."
ssh -i "${SSH_KEY}" "${SERVER}" "mkdir -p ${REMOTE_APP_DIR}/static ${REMOTE_APP_DIR}/assets"
scp -i "${SSH_KEY}" -r static/. "${SERVER}:${REMOTE_APP_DIR}/static/"
scp -i "${SSH_KEY}" -r assets/. "${SERVER}:${REMOTE_APP_DIR}/assets/"

echo "==> Przesyłanie pliku środowiska produkcyjnego..."
scp -i "${SSH_KEY}" "${SCRIPT_DIR}/.env.prod" "${SERVER}:${REMOTE_APP_DIR}/.env"
rm -f "${SCRIPT_DIR}/.env.prod"

echo "==> Instalacja binarki i restart usługi..."
ssh -i "${SSH_KEY}" "${SERVER}" bash -s <<EOF
set -euo pipefail
sudo install -m 755 /tmp/${BINARY_NAME} /usr/local/bin/${BINARY_NAME}
rm -f /tmp/${BINARY_NAME}
sudo chown -R ubuntu:ubuntu ${REMOTE_APP_DIR}

SERVICE_FILE=/etc/systemd/system/sumit.service
ENV_LINE='EnvironmentFile=-/opt/sumit/.env'
if [[ -f "\${SERVICE_FILE}" ]]; then
  if ! sudo grep -qF "\${ENV_LINE}" "\${SERVICE_FILE}"; then
    echo "==> Dodawanie EnvironmentFile do sumit.service..."
    sudo sed -i '/^\[Service\]/a EnvironmentFile=-/opt/sumit/.env' "\${SERVICE_FILE}"
    sudo systemctl daemon-reload
  fi
fi

if systemctl is-active --quiet sumit; then
  sudo systemctl restart sumit
else
  echo "Usługa sumit nie jest jeszcze włączona — uruchom jednorazową konfigurację z DEPLOY.md"
fi
EOF

echo "==> Wdrożenie zakończone."
