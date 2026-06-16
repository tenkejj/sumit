# Deploy

Instructions for maintainers. End users only need the live URL from [README.md](README.md).

## Config

Copy the example file and fill in your values:

```bash
cp deploy.env.example deploy.env
```

`deploy.env` is gitignored — never commit keys or server addresses there.

| Variable | Example |
|---|---|
| `SERVER` | `ubuntu@203.0.113.10` |
| `SSH_KEY` | `/home/you/.ssh/id_ed25519` |

On Windows (Git Bash), key paths look like `/c/Users/you/.ssh/key`.

You can also export variables in the shell instead of using `deploy.env`.

## One-time server setup

Ubuntu 24.04, user `ubuntu`. Go is **not** required on the server.

```bash
source deploy.env   # or: export SERVER=... SSH_KEY=...

ssh -i "${SSH_KEY}" "${SERVER}"

# on the server:
sudo mkdir -p /opt/sumit/static /opt/sumit/assets
sudo chown -R ubuntu:ubuntu /opt/sumit
exit

scp -i "${SSH_KEY}" sumit.service "${SERVER}:/tmp/sumit.service"
ssh -i "${SSH_KEY}" "${SERVER}" \
  'sudo mv /tmp/sumit.service /etc/systemd/system/ && sudo systemctl daemon-reload && sudo systemctl enable --now sumit'

# optional — open HTTP port:
ssh -i "${SSH_KEY}" "${SERVER}" 'sudo ufw allow 8080/tcp'
```

The app listens on port **8080**.

## Deploy

From the project root:

```bash
./deploy.sh
```

The script cross-compiles for `linux/amd64`, uploads the binary and `static/` + `assets/` via `scp`, installs to `/usr/local/bin/sumit`, and restarts the `sumit` systemd unit.

## Verify

```bash
source deploy.env

ssh -i "${SSH_KEY}" "${SERVER}" 'sudo systemctl status sumit'
ssh -i "${SSH_KEY}" "${SERVER}" 'journalctl -u sumit -f'
```

---

# Wdrożenie

Instrukcja dla maintainera. Użytkownicy końcowi potrzebują tylko linku z [README.md](README.md).

## Konfiguracja

```bash
cp deploy.env.example deploy.env
```

Plik `deploy.env` jest w `.gitignore` — nie commituj tam kluczy ani adresów serwera.

| Zmienna | Przykład |
|---|---|
| `SERVER` | `ubuntu@203.0.113.10` |
| `SSH_KEY` | `/home/you/.ssh/id_ed25519` |

Na Windows w Git Bash ścieżki wyglądają jak `/c/Users/you/.ssh/key`.

Zmienne możesz też ustawić w shellu zamiast trzymać je w `deploy.env`.

## Jednorazowa konfiguracja serwera

Ubuntu 24.04, użytkownik `ubuntu`. Go na serwerze **nie jest potrzebne**.

```bash
source deploy.env   # lub: export SERVER=... SSH_KEY=...

ssh -i "${SSH_KEY}" "${SERVER}"

# na serwerze:
sudo mkdir -p /opt/sumit/static /opt/sumit/assets
sudo chown -R ubuntu:ubuntu /opt/sumit
exit

scp -i "${SSH_KEY}" sumit.service "${SERVER}:/tmp/sumit.service"
ssh -i "${SSH_KEY}" "${SERVER}" \
  'sudo mv /tmp/sumit.service /etc/systemd/system/ && sudo systemctl daemon-reload && sudo systemctl enable --now sumit'

# opcjonalnie — otwórz port HTTP:
ssh -i "${SSH_KEY}" "${SERVER}" 'sudo ufw allow 8080/tcp'
```

Aplikacja nasłuchuje na porcie **8080**.

## Wdrożenie

Z katalogu głównego projektu:

```bash
./deploy.sh
```

Skrypt kompiluje pod `linux/amd64`, wysyła binarkę oraz `static/` i `assets/` przez `scp`, instaluje do `/usr/local/bin/sumit` i restartuje usługę `sumit`.

## Sprawdzenie

```bash
source deploy.env

ssh -i "${SSH_KEY}" "${SERVER}" 'sudo systemctl status sumit'
ssh -i "${SSH_KEY}" "${SERVER}" 'journalctl -u sumit -f'
```
