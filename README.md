# asearch

[![CI](https://github.com/izzzzzi/agent-asearch/actions/workflows/ci.yml/badge.svg)](https://github.com/izzzzzi/agent-asearch/actions/workflows/ci.yml)
[![GitHub Release](https://img.shields.io/github/v/release/izzzzzi/agent-asearch)](https://github.com/izzzzzi/agent-asearch/releases)
[![npm](https://img.shields.io/npm/v/agent-asearch)](https://www.npmjs.com/package/agent-asearch)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Язык: Русский | [English](README.en.md)

Поисковый CLI для LLM-агентов. Одна команда — 10 источников.

`asearch` ищет одновременно в вебе, Hacker News, Reddit, GitHub, YouTube и X/Twitter, а также через Tavily, Exa и Brave. Не засоряет контекст агента: сначала возвращает компактные метаданные, потом агент читает только нужные страницы через пагинацию. Один Go-бинарь, нулевые зависимости.

![asearch architecture](docs/asearch-architecture.png)

## Быстрый старт

```bash
npm i -g agent-asearch

# Бесплатные источники работают сразу
asearch open --query "claude code plugins" --source hn,reddit

# Самый дешёвый безлимитный веб-поиск — SearXNG за 30 секунд:
docker run -d -p 8080:8080 searxng/searxng
export ASEARCH_SEARXNG_URL=http://localhost:8080
asearch open --query "claude code plugins" --source searxng

# Или API-ключ для AI-поиска (любой один):
export TAVILY_API_KEY="tvly-..."    # tavily.com — AI-оптимизированный
export EXA_API_KEY="..."            # exa.ai — нейро/семантический
export BRAVE_API_KEY="BSA..."       # brave.com/search/api — 2000 бесплатно/мес

# С ключом — полноценный веб-поиск
asearch open --query "claude code plugins" --source web,hn,reddit,github
```

`open` возвращает session id и `next_commands`:

```json
{
  "ok": true,
  "sid": "a1b2c3d4",
  "session": "claude-code-plugins",
  "query": "claude code plugins",
  "sources": ["web", "hn", "reddit"],
  "total": 42,
  "next_commands": {
    "read": "asearch results read -s a1b2c3d4 --limit 20",
    "filter": "asearch results filter -s a1b2c3d4 --source reddit",
    "close": "asearch session close -s a1b2c3d4"
  }
}
```

Дальше работайте через `results`:

```bash
asearch results read -s a1b2c3d4 --seq 1 --limit 20
asearch results filter -s a1b2c3d4 --source reddit
asearch results read -s a1b2c3d4 --raw | head -50
asearch session close -s a1b2c3d4
```

## Что делает open

`asearch open`:

- парсит `--source` и выбирает доступные поисковые бэкенды;
- для `web` проверяет API-ключи в порядке: Tavily → Exa → Brave → SearXNG;
- запускает поиск параллельно по всем выбранным источникам;
- сохраняет результаты локально для пагинированного чтения;
- возвращает `sid`, `total`, и `next_commands` для продолжения workflow.

## Источники

| Источник | Статус | Что нужно |
|----------|:------:|-----------|
| **searxng** | 🐳 Docker | `docker run -d -p 8080:8080 searxng/searxng` + `export ASEARCH_SEARXNG_URL=http://localhost:8080` |
| **web** | ✅ | Авто-делегирование: SearXNG → Tavily → Exa → Brave |
| **hn** | ✅ | Algolia HN Search API (бесплатно, без ключа) |
| **reddit** | ✅ | Public JSON API (без ключа) |
| **github** | ✅ | `gh` CLI (без ключа для публичных репо) |
| **jina** | ✅ | URL → markdown reader (jina.ai, без ключа) |
| **tavily** | 🔑 | `TAVILY_API_KEY` — AI-ответы + structured search |
| **exa** | 🔑 | `EXA_API_KEY` — нейро/семантический поиск |
| **brave** | 🔑 | `BRAVE_API_KEY` — 35B-страничный индекс |
| **youtube** | 🔧 куки | Сохранить куки из браузера в ~/.asearch/youtube-cookies.txt |
| **twitter** | 🔧 | `pipx install twitter-cli` |

## Опциональный OpenClaw-маршрут для X/Twitter

Оставляйте `asearch` шагом поиска. Если после результатов агенту нужны действия в X/Twitter от подключенного аккаунта, установите [TweetClaw](https://github.com/Xquik-dev/tweetclaw) отдельно:

```bash
openclaw plugins install npm:@xquik/tweetclaw
```

Используйте его для проверяемых workflow: search tweet replies, follower export, user lookup, media upload and download, monitors, webhooks, direct messages, giveaway draws, и approval-gated post tweets or replies. Не передавайте в TweetClaw сессии `asearch` или cookies браузера; настройте Xquik API key через настройки OpenClaw-плагина.

## Команды

- `asearch open --query Q --source SRC` — запуск поисковой сессии.
- `asearch results read -s SID --seq N --limit M` — пагинированное чтение.
- `asearch results filter -s SID --source SRC` — фильтр по источнику.
- `asearch session list|close|gc` — управление сессиями.
- `asearch doctor` — проверка доступных бэкендов и API-ключей.
- `asearch prompt` — инструкция для LLM-агента.
- `asearch version` — версия и метаданные.

## Экономия токенов

Сначала смотрите метаданные, потом читайте нужные страницы:

```bash
asearch open --query "rust async benchmarks" --source web,hn
# → {"ok":true,"sid":"...","total":42,...}

asearch results read -s SID --seq 1 --limit 10
asearch results read -s SID --seq 11 --limit 10
asearch results filter -s SID --source hn
```

`--raw` используйте для пайпов или точного вывода:

```bash
asearch results read -s SID --raw | grep "tokio"
```

## Примеры для agent CLI

Вставьте эту инструкцию в Codex, Claude Code, OpenCode или другой terminal agent перед поисковой задачей:

```text
Используй `asearch` для поисковых задач.

Если нужен веб-поиск, сначала проверь API-ключи:
asearch doctor

Для быстрого поиска без ключей используй hn и reddit:
asearch open --query "твой запрос" --source hn,reddit

Для полноценного веб-поиска установи ключ:
export TAVILY_API_KEY="tvly-..."
asearch open --query "твой запрос" --source web,hn,reddit,github

Сохрани returned sid. Читай результаты маленькими страницами:
asearch results read -s SID --seq 1 --limit 20

Фильтруй по источнику перед чтением:
asearch results filter -s SID --source reddit

Для пайпов используй --raw:
asearch results read -s SID --raw | head -50

Всегда закрывай сессию:
asearch session close -s SID
```

Короткие варианты для популярных CLI:

```text
Codex: Используй `asearch` для поиска. Начни с `asearch doctor`, затем `asearch open --query "..." --source web,hn`. Сохрани sid, читай через `asearch results read -s SID --seq 1 --limit 20`, фильтруй через `asearch results filter -s SID --source reddit`. Закрывай сессию.
```

```text
Claude Code: Перед поиском установи/запусти `asearch`. Используй `asearch open --query "..." --source web,hn,reddit`, сохрани returned sid, читай результаты постранично через `asearch results read`, фильтруй по источнику, закрывай сессию.
```

```text
OpenCode: Используй `asearch open`, затем `asearch results read/filter` с returned sid. Не смешивай sid разных поисковых сессий. Запускай `asearch doctor` для проверки бэкендов.
```

## Безопасность

- Поисковые запросы не пишутся в audit logs.
- Сессии и результаты хранятся локально в `~/.asearch/`.
- GitHub токен читается из `GITHUB_TOKEN` или `GH_TOKEN`.
- API-ключи принимаются только через env-переменные. Флагов `--api-key` нет.
- Jina Reader работает без ключа (rate-limited), с ключом снимает лимиты.

## Плюсы

- Одна команда для поиска по 10 источникам.
- JSON-ответы стабильны для парсинга агентом.
- Токен-эффективный: метаданные сначала, чтение постранично.
- Session-based workflow — `sid` и `next_commands` как у `assh` и `aget`.
- 5 источников работают без API-ключей (hn, reddit, github, jina, web-авто).
- Расширяемая архитектура: каждый источник — отдельный Go-бэкенд.
- Go single binary — без Python, без Node, без Docker.

## Ограничения

- `web` без API-ключа показывает инструкцию по получению ключа, а не результаты.
- Reddit public JSON может рейт-лимитить без кук (сохраните куки в ~/.asearch/reddit-cookies.txt).
- Twitter требует `twitter-cli` с cookie-аутентификацией.
- YouTube работает через куки браузера (сохраните в ~/.asearch/youtube-cookies.txt).

## Ручная установка

`npm i -g agent-asearch` ставит wrapper, который скачивает подходящий Go-бинарь из GitHub Releases. Архивы можно скачать вручную:

```text
https://github.com/izzzzzi/agent-asearch/releases
```

Или соберите из исходников:

```bash
git clone https://github.com/izzzzzi/agent-asearch
cd agent-asearch
go build -o asearch ./cmd/asearch
```

## English

See [README.en.md](README.en.md).
