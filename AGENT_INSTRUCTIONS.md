# AGENT_INSTRUCTIONS.md — agent-asearch

Инструкция для LLM-агентов по использованию `asearch`.

→ Краткая версия: `asearch prompt`
→ Проверка доступного: `asearch doctor`

---

## Быстрый старт (30 секунд до первого результата)

```bash
# Ничего не нужно устанавливать — HN работает сразу
asearch open --query "claude code plugins" --source hn
asearch results read -s SID --seq 1 --limit 20
asearch session close -s SID
```

---

## Установка и активация каждого провайдера

### Zero-config — работают сразу

| Провайдер | Команда |
|-----------|---------|
| **hn** | `asearch open --query "..." --source hn` |
| **github** | `asearch open --query "..." --source github` (нужен `gh` на PATH) |
| **jina** | `asearch open --query "..." --source jina` (URL → markdown) |
| **reddit** | Сохрани куки: `~/.asearch/reddit-cookies.txt` (Netscape-формат из браузера) |

### Установка через pipx/brew (одна команда)

| Провайдер | Что сделать |
|-----------|-------------|
| **youtube** | Сохранить куки `~/.asearch/youtube-cookies.txt` (Netscape из браузера) |
| **twitter** | `pipx install twitter-cli` → `twitter login` (откроет браузер один раз) |

### Self-hosted — свой сервер, без лимитов

```bash
# SearXNG — агрегирует Google, Bing, DDG, Wikipedia + 70 движков
docker run -d -p 8080:8080 searxng/searxng
export ASEARCH_SEARXNG_URL=http://localhost:8080
asearch open --query "..." --source searxng
```

### API-ключи — любой один, бесплатные тиры у всех

| Провайдер | Env var | Где взять | Бесплатно |
|-----------|---------|-----------|:---------:|
| **Tavily** | `TAVILY_API_KEY="tvly-..."` | tavily.com | ✅ |
| **Perplexity** | `PERPLEXITY_API_KEY="pplx-..."` | docs.perplexity.ai | ✅ |
| **Exa** | `EXA_API_KEY="..."` | exa.ai | ✅ |
| **Brave** | `BRAVE_API_KEY="BSA..."` | brave.com/search/api | 2000/мес |
| **Serper** | `SERPER_API_KEY="..."` | serper.dev | 2500/мес |
| **SerpAPI** | `SERPAPI_API_KEY="..."` | serpapi.com | 100/мес |
| **You.com** | `YOU_API_KEY="..."` | you.com/api | ✅ |
| **Firecrawl** | `FIRECRAWL_API_KEY="fc-..."` | firecrawl.dev | 500/мес |
| **Parallel** | `PARALLEL_API_KEY="..."` | parallel.ai | ✅ |

После установки любого ключа `--source web` автоматически делегирует в него.

---

## Основные команды

| Команда | Описание |
|---------|----------|
| `asearch doctor` | Проверить все 16 провайдеров — какие готовы, какие нет |
| `asearch open --query Q --source SRC` | Запустить поисковую сессию |
| `asearch results read -s SID --seq N --limit M` | Пагинированное чтение |
| `asearch results filter -s SID --source SRC` | Фильтр по источнику |
| `asearch session list` | Список сессий |
| `asearch session close -s SID` | Закрыть сессию |
| `asearch session gc` | Очистить старые |

---

## Web delegation chain

`--source web` пробует провайдеров по порядку, первый доступный используется:

```
SearXNG → Tavily → Perplexity → Exa → Brave → Serper → SerpAPI → You → Firecrawl → Parallel
```

---

## JSON-контракт

### Успех (`open`)
```json
{
  "ok": true, "sid": "a1b2c3d4", "total": 42,
  "next_commands": {
    "read": "asearch results read -s a1b2c3d4 --limit 20",
    "filter": "asearch results filter -s a1b2c3d4 --source reddit",
    "close": "asearch session close -s a1b2c3d4"
  }
}
```

### Ошибка
```json
{"ok":false,"code":"invalid_args","message":"query required","hint":"use --query \"...\""}
```

---

## Экономия токенов

1. `open` → смотри `total`, оцени объём
2. `results filter --source X` → сузь выборку
3. `results read --seq 1 --limit 20` → читай маленькими страницами
4. `--raw` для grep/head

---

## Пример workflow

```bash
# 1. Проверить что доступно
asearch doctor

# 2. Если ничего нет — поднять SearXNG (30 секунд)
docker run -d -p 8080:8080 searxng/searxng
export ASEARCH_SEARXNG_URL=http://localhost:8080

# 3. Поиск
asearch open --query "best AI coding tools 2026" --source searxng,hn,github

# 4. Чтение
asearch results read -s SID --seq 1 --limit 20

# 5. Фильтр
asearch results filter -s SID --source hn

# 6. Закрыть
asearch session close -s SID
```
