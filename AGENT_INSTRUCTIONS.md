# AGENT_INSTRUCTIONS.md — agent-asearch

Инструкция для LLM-агентов по использованию `asearch` — CLI для мульти-источникового поиска.

## Быстрый старт

```bash
# Проверить доступные бэкенды
asearch doctor

# Поиск по бесплатным источникам (без API-ключей)
asearch open --query "claude code plugins" --source hn,reddit

# Сохранить returned sid и прочитать первую страницу
asearch results read -s SID --seq 1 --limit 20

# Отфильтровать результаты по источнику
asearch results filter -s SID --source reddit

# Закрыть сессию
asearch session close -s SID
```

## Основные команды

| Команда | Описание |
|---------|----------|
| `asearch open --query Q --source SRC...` | Запустить поисковую сессию |
| `asearch results read -s SID --seq N --limit M` | Чтение результатов постранично |
| `asearch results filter -s SID --source SRC` | Фильтр по источнику |
| `asearch session list` | Список активных сессий |
| `asearch session close -s SID` | Закрыть сессию |
| `asearch session gc` | Очистить старые закрытые сессии |
| `asearch doctor` | Проверить доступные бэкенды и API-ключи |
| `asearch prompt` | Инструкция для агента |
| `asearch version` | Версия и метаданные |

## Источники

### Без настройки (zero-config)

- **hn** — Algolia Hacker News Search API (всегда работает)
- **reddit** — Public JSON API (может рейт-лимитить без кук)
- **github** — `gh` CLI (для публичных репозиториев)
- **jina** — URL → markdown reader (jina.ai)
- **web** — авто-делегирование: если есть Tavily/Exa/Brave ключ — использует его, иначе показывает инструкцию

### Требуют API-ключа

- **tavily** — `export TAVILY_API_KEY="tvly-..."` (tavily.com, бесплатный тир)
- **exa** — `export EXA_API_KEY="..."` (exa.ai, нейро-поиск)
- **brave** — `export BRAVE_API_KEY="BSA..."` (brave.com/search/api, 2000 бесплатно/мес)

### Требуют установки

- **youtube** — `brew install yt-dlp`
- **twitter** — `pipx install twitter-cli`

## JSON-контракт

### Успех (`open`)
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

### Результаты (`results read`)
```json
{
  "ok": true,
  "sid": "a1b2c3d4",
  "total": 42,
  "seq": 1,
  "limit": 20,
  "count": 20,
  "results": {
    "items": [
      {
        "seq": 1,
        "source": "web",
        "title": "Best Claude Code Plugins 2026",
        "url": "https://example.com",
        "snippet": "...",
        "date": "2026-05-01",
        "score": 0.95,
        "engagement": "score: 0.95"
      }
    ]
  },
  "next_commands": {
    "next": "asearch results read -s a1b2c3d4 --seq 21 --limit 20",
    "prev": "asearch results read -s a1b2c3d4 --seq 1 --limit 20",
    "close": "asearch session close -s a1b2c3d4"
  }
}
```

### Ошибка
```json
{"ok":false,"code":"invalid_args","message":"query required","hint":"use --query \"...\""}
```

## Экономия токенов

1. **Сначала метаданные.** `open` возвращает `total` и `sources` — оцените объём до чтения.
2. **Маленькие страницы.** `--limit 20` даёт компактный JSON, который агент легко парсит.
3. **Фильтруйте до чтения.** `results filter --source reddit` сужает выборку.
4. **`--raw` для пайпов.** `results read --raw | grep "pattern"` — без JSON-обёртки.

```bash
# Плохо: читать всё сразу
asearch results read -s SID --seq 1 --limit 200

# Хорошо: метаданные → фильтр → страница
asearch open --query "topic" --source web,hn,reddit
asearch results filter -s SID --source hn
asearch results read -s SID --seq 1 --limit 10
```

## Пример workflow для агента

```bash
# 1. Проверить доступность бэкендов
asearch doctor

# 2. Запустить поиск
asearch open --query "best AI coding tools 2026" --source web,hn,reddit,github

# 3. Прочитать первую страницу
asearch results read -s a1b2c3d4 --seq 1 --limit 20

# 4. Отфильтровать Reddit и прочитать отдельно
asearch results filter -s a1b2c3d4 --source reddit

# 5. Для пайпа — raw
asearch results read -s a1b2c3d4 --raw | grep "claude"

# 6. Закрыть сессию
asearch session close -s a1b2c3d4
```

## Переменные окружения

- `TAVILY_API_KEY` — ключ Tavily (tavily.com).
- `EXA_API_KEY` — ключ Exa (exa.ai).
- `BRAVE_API_KEY` — ключ Brave Search (brave.com/search/api).
- `JINA_API_KEY` — ключ Jina Reader (jina.ai, опционально).
- `GITHUB_TOKEN` / `GH_TOKEN` — токен GitHub API.
- `ASEARCH_STATE_DIR` — каталог для сессий и результатов (по умолчанию `~/.asearch`).
- `ASEARCH_SEARXNG_URL` — URL self-hosted SearXNG.
- `AGENT_ASEARCH_SKIP_DOWNLOAD=1` — пропустить скачивание native-бинаря в npm `postinstall`.

## Безопасность

- Поисковые запросы не логируются в открытом виде.
- Сессии и результаты хранятся локально в `~/.asearch/` с permissions 0700.
- API-ключи принимаются только через env-переменные. Флага `--api-key` нет.
- GitHub токен читается из `GITHUB_TOKEN` или `GH_TOKEN`, никогда не печатается.
- Jina Reader работает без ключа, с ключом повышает rate limits.

## Плюсы

- Мульти-источниковый поиск одной командой.
- JSON-ответы стабильны для парсинга агентом.
- Токен-эффективный: метаданные → фильтр → постраничное чтение.
- Session-based workflow с `sid` и `next_commands`.
- 5 источников работают без настройки.
- Расширяемая архитектура (каждый источник — отдельный Go-бэкенд).
- Go single binary — без зависимостей.

## Ограничения

- `web` без API-ключа показывает инструкцию, а не результаты.
- Reddit public JSON может рейт-лимитить без кук.
- Twitter требует cookie-аутентификацию через `twitter-cli`.
- YouTube требует `yt-dlp` (неинтерактивный режим).
