# Передача Nemotron другой сессии: запуск и интеграция

Подготовлено 2026-09-23. Это инструкция для **существующей настроенной VM**.
Последняя проверка health: 12:32:53 UTC; ASR → LLM smoke test прошёл.
При новой сессии сначала проверить текущее состояние командами ниже.

## 1. Что уже есть и как войти

- VM: `nemotron-voice-agent-402b8f`, Brev / AWS Columbus, RTX PRO 6000 Blackwell Server 96 GB.
- [Jupyter и сохранённый notebook](https://jupyter-vt4eo8rll.gobrev.dev/lab/tree/brev_launchable.ipynb).
- [Готовый voice UI](https://nemotron-voice-agent-vt4eo8rll.gobrev.dev/).
- Оба HTTPS URL защищены входом владельца в Brev. Это не анонимные API для Vercel.
- Проверенный способ управления — Jupyter. Открыть в нём **File → New → Terminal**
  либо **Launcher → Terminal** и выполнять Bash-команды ниже в терминале VM.
- Если другая сессия использует отдельный браузер, владелец должен войти в NVIDIA/Brev
  там. Пароль, cookies и коды подтверждения в чат или код переносить не требуется.
- SSH в этой работе не настраивался и не проверялся. Публичный IP из карточки
  `3.14.153.223`, SSH gateway `global.prd.ga.run.brev.nvidia.com:36400` — справочные
  данные; адрес сам по себе не даёт доступ без авторизованного SSH-ключа.

Два разных файловых пространства:

| Где | Каталог |
|---|---|
| Локальный проект Codex | `/home/agent/projects/hack-c52f5e57-butterfly` |
| NVIDIA Blueprint на VM Brev | `/home/ubuntu/nemotron-voice-agent` |

Ни одна команда с `127.0.0.1` ниже не обращается к Brev с вашего ноутбука.
Она работает **на VM** либо через явно настроенный туннель, которого пока нет.

## 2. Проверка и возобновление запуска

В новом Bash-терминале **VM**:

```bash
cd /home/ubuntu/nemotron-voice-agent
git rev-parse HEAD
nvidia-smi
test -f .env

# Используем настройки уже подготовленного .env, без случайных shell overrides.
unset EXAMPLE_SELECTION PIPELINE_TLS PIPELINE_APP_PORT TRANSPORT_SELECTION
NEMOTRON_COMPOSE=(docker compose --env-file .env --profile multilingual-assistant/single-gpu --profile turn)
"${NEMOTRON_COMPOSE[@]}" config -q
"${NEMOTRON_COMPOSE[@]}" ps --all
```

Проверенный commit: `13801d9963aea20ac663318060824a78b28809fa`.
Не обновлять upstream перед интеграцией без отдельной причины. Веса и образы уже
загружены; повторная установка Go, Docker, моделей или новая VM не нужны.

Ожидаются четыре сервиса:

```text
multilingual-assistant-single-gpu
nemo-speech-multilingual
nvidia-llm-vllm-lightning
coturn
```

Если они уже работают, переходить к проверке API. Если контейнеры существуют,
но остановлены:

```bash
"${NEMOTRON_COMPOSE[@]}" start
```

Если контейнеры отсутствуют, но `.env`, образы и кеш весов сохранены:

```bash
"${NEMOTRON_COMPOSE[@]}" up -d --no-recreate --no-build --pull never
```

Эти команды возобновления приведены по закреплённой конфигурации; отдельный
restart-тест после успешного первого запуска не проводился. Первичная загрузка
LLM включает прогрев/компиляцию: статус контейнера `Up` ещё не означает готовность.
Если сама VM остановлена, сначала запустить её в Brev, затем выполнить проверки.

Не выполнять notebook **Run All** ради подключения: его Deploy-ячейка останавливает
и пересоздаёт стек. Не использовать `down -v`, удаление volume/cache или
`RESET_VOLUMES=True`. Текущий `.env` содержит приватные credentials и имеет режим
`0600`; не печатать его и не переносить в Git. Для диагностики Compose использовать
`config -q`, поскольку полный `config` может раскрыть интерполированные секреты.

## 3. Проверка готовности

В том же терминале VM:

```bash
curl --noproxy '*' -fsS --max-time 10 -o /dev/null -w 'APP HTTP %{http_code}\n' http://127.0.0.1:7860/health
curl --noproxy '*' -fsS --max-time 10 -o /dev/null -w 'LLM HTTP %{http_code}\n' http://127.0.0.1:18000/health
curl --noproxy '*' -fsS --max-time 10 http://127.0.0.1:18000/v1/models
"${NEMOTRON_COMPOSE[@]}" logs --tail=60 nvidia-llm-vllm-lightning
```

APP и LLM должны вернуть **200**, а models — ID
`nvidia/nemotron-3.5-lightning-30b-a3b`.
ASR проверяется следующим smoke test через реальный Riva RPC, а не через HTTP.
Отсутствие JSON-тела у успешного LLM `/health` допустимо.

## 4. Повторить уже успешный тест

На хосте VM сохранены `/home/ubuntu/smoke-final.py` и
`/home/ubuntu/meeting-2.mp3`. Копия того же клиента есть в локальном проекте:
[`scripts/nemotron_smoke.py`](../../scripts/nemotron_smoke.py).
Его SHA256: `fe08a1e9c37cc7611acc2a97ad9f492e265fbc21f7c9999605c3b55b83d6693d`.

В терминале **VM**, после определения `NEMOTRON_COMPOSE` из пункта 2:

```bash
sha256sum /home/ubuntu/smoke-final.py
"${NEMOTRON_COMPOSE[@]}" cp /home/ubuntu/smoke-final.py multilingual-assistant-single-gpu:/tmp/smoke.py
"${NEMOTRON_COMPOSE[@]}" cp /home/ubuntu/meeting-2.mp3 multilingual-assistant-single-gpu:/tmp/meeting-2.mp3

NEMOTRON_RUN_ID="meeting2-$(date -u +%Y%m%dT%H%M%S)-$$"
"${NEMOTRON_COMPOSE[@]}" exec -T multilingual-assistant-single-gpu \
  uv run --frozen python /tmp/smoke.py /tmp/meeting-2.mp3 \
  --container-network --seconds 30 --out "/tmp/$NEMOTRON_RUN_ID"

mkdir -p /home/ubuntu/butterfly-test-results
chmod 700 /home/ubuntu/butterfly-test-results
"${NEMOTRON_COMPOSE[@]}" cp \
  "multilingual-assistant-single-gpu:/tmp/$NEMOTRON_RUN_ID" \
  /home/ubuntu/butterfly-test-results/
```

Клиент использует зависимости существующего app-контейнера; ставить Python-пакеты
на хост для этого не нужно. Каждый запуск требует нового каталога результата.
`--asr-only` проверит только ASR; `--start 30 --seconds 30` выберет следующий фрагмент.
Предел клиента — 60 секунд за запуск: это диагностический клиент, не готовый
обработчик полной встречи.

Подтверждённый результат первого полного smoke test:

```json
{"status":"passed","audio_seconds":30.0,"asr_seconds":0.9295554129998891,"llm_seconds":2.03318032199968,"transcript_characters":455,"actions_count":1,"accuracy_evaluated":false,"diarization_evaluated":false}
```

Первый результат уже сохранён вне контейнера:
`/home/ubuntu/butterfly-test-results/meeting2-smoke-1/`.
Внутри `clip.wav`, `clip.json`, `asr.json`, `actions.json`, `result.json`;
рядом `meeting2-smoke-1-manifest.json`. Каталог `0700`, файлы `0600`.
При неудаче клиент создаёт `failure.json`; после пересоздания контейнера сначала
снова скопировать клиент и аудио с хоста. Не переносить аудио и стенограммы в Git.

## 5. Подключить backend к моделям

Это **предлагаемые имена переменных нашего адаптера**, а не параметры upstream.
Приложение должно явно читать их и использовать указанные адреса.

Backend-процесс на **хосте Brev VM**:

```dotenv
RIVA_ASR_ENDPOINT=127.0.0.1:50051
ASR_LANGUAGE=ru-RU
LOCAL_LLM_BASE_URL=http://127.0.0.1:18000/v1
LOCAL_LLM_MODEL=nvidia/nemotron-3.5-lightning-30b-a3b
```

Backend-контейнер **в одной Docker-сети с этим Compose-стеком**:

```dotenv
RIVA_ASR_ENDPOINT=nemo-speech-multilingual:50051
ASR_LANGUAGE=ru-RU
LOCAL_LLM_BASE_URL=http://nvidia-llm-vllm:8000/v1
LOCAL_LLM_MODEL=nvidia/nemotron-3.5-lightning-30b-a3b
```

Порт LLM внутри Docker — **8000**, на хосте — **18000**. `localhost` внутри нового
контейнера указывает на него самого. DNS-имена выше доступны только при подключении
к сети Blueprint. Узнать имя сети без просмотра секретов:

```bash
NEMOTRON_APP_ID=$("${NEMOTRON_COMPOSE[@]}" ps -q multilingual-assistant-single-gpu)
docker inspect --format '{{range $name, $config := .NetworkSettings.Networks}}{{println $name}}{{end}}' "$NEMOTRON_APP_ID"
```

В отдельном Compose backend объявить найденную сеть как `external: true`, задать
её точное `name` и подключить backend к ней. Не предполагать, что две разные
Compose-конфигурации автоматически имеют общую сеть.

**ASR:** Riva gRPC, `riva.client.ASRService` и `Recognize`. Проверенный вход — raw
PCM16 little-endian, mono, 16000 Hz, `language_code="ru-RU"`. WAV нужно декодировать
и передавать PCM frames, а не MP3-байты или WAV-заголовок. Рабочая реализация уже
есть в функциях `prepare_clip()` и `transcribe()` клиента. Он запрашивает каталог
`GetRivaSpeechRecognitionConfig`, выбирает объявленную модель для `ru-RU` и
запрашивает word timestamps. Разделения по спикерам этот вызов не добавляет.
Длинные файлы требуют отдельно реализованной потоковой или сегментной обработки.

**LLM:** OpenAI-compatible протокол означает формат HTTP, а не обращение к OpenAI.
Проверенный запрос: `POST /v1/chat/completions`, `temperature=0`,
`response_format={"type":"json_object"}`,
`chat_template_kwargs={"enable_thinking":false}`. В smoke использован лимит
1536 output tokens; для полной встречи это не универсальный лимит. Проверять
`finish_reason`, схему JSON и цитаты. Рабочая реализация — `extract_actions()`.
Локальный endpoint в проверенном стеке не требовал Authorization header или
OpenAI API key. При использовании SDK обязательно передать явный `base_url`;
не полагаться на его облачный адрес по умолчанию.

Контракт проверенного извлечения:

```json
{
  "summary": "Краткое содержание",
  "actions": [{
    "task": "Подготовить отчёт",
    "assignee": null,
    "deadline_raw": null,
    "evidence_quote": "точная подстрока исходной стенограммы"
  }]
}
```

Это описание структуры, не результат записи. Ответственный и срок допускают
`null`; дата встречи нужна для нормализации относительных сроков. Говорящий и
ответственный по поручению — разные поля. Схема продукта может быть расширена,
но должна сохранять проверяемую связь с исходным текстом.

## 6. Что интегрировать в проекте и Vercel

Текущий voice UI — отдельный NVIDIA demo. API загрузки встреч и протоколов в нём
не реализовано. Рекомендуемая схема нашего приложения:

```text
Browser / Vercel frontend
  → authenticated HTTPS meeting backend on Brev
  → local Riva ASR + local Nemotron vLLM
  → private transcript / actions / PDF / DOCX
```

Backend нужно разместить на этой VM, добавить приём файла, очередь/статус задания,
обработку и экспорт. Сначала проверить локальную цепочку. Затем настроить отдельный
HTTPS endpoint backend с прикладной авторизацией и CORS для точного Vercel origin.
Публичный URL frontend на Vercel не должен содержать ключи моделей или постоянный
backend secret; `VITE_*` / `NEXT_PUBLIC_*` доступны посетителю сайта.

Текущий Brev Secure Link требует owner login; копирование его URL в server-side
Vercel env не создаёт machine-to-machine доступ. Для открытой демки потребуется
отдельно согласованная настройка внешнего доступа к backend. Нельзя считать
успешным deployment лишь потому, что страница Vercel открывается: проверить upload
из браузера, CORS/preflight, статус обработки и скачивание результата.
Порты сырых моделей `18000` и `50051` публично для интеграции не открывать.

Функции для следующей сессии: full meeting upload → ASR → diarization → assignments
and deadlines → review → PDF/DOCX, затем Fly, Telegram и финальный Vercel.
Полный продукт, точность, пять спикеров, KK/mixed и offline/on-premise запуск пока
не подтверждены. В stock voice UI сейчас есть `de-DE`, `en-US`, `es-ES`, `fr-FR`,
`it-IT`; русский работает в проверенном ASR/LLM API-тесте, а не в его voice selector.

## 7. Файлы, на которые опираться

- [Текущий runtime и результаты](nemotron-runtime-report.md).
- [Адреса, статусы, hashes и параметры интеграции](runtime-handoff.json).
- [Диагностический клиент](../../scripts/nemotron_smoke.py).
- [Руководство первоначального запуска](brev-gpu.md).
- Локальные fixtures: `/home/agent/projects/hack-c52f5e57-butterfly/data/test/meeting-minutes/`.
  Это gitignored-данные; в другой checkout сами по себе не появятся.
- На VM загружена вторая запись; первую и эталоны для полного сравнения ещё нужно
  перенести разрешённым способом из локальных fixtures.

Предыдущий preparation-report описывает состояние до запуска и сохранён как история.
Файлы передачи подготовлены в текущем workspace; наличие их в другой сессии/checkout
нужно проверить. Для этого handoff отдельный commit/push не выполнялся.
