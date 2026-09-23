# Butterfly → существующий Nemotron на Brev

Интеграция добавлена в **Go HTTP backend**. Перед запуском прочитайте
[handoff](nemotron-integration-handoff.md). Модели и notebook повторно не запускаются.
Сведения о smoke исходной сессии в `runtime-handoff.json` — историческое измерение,
а не результат проверки нового backend.

## Что запускается

Go обращается к локальному Nemotron `/v1/chat/completions` с `enable_thinking=false`,
`response_format=json_object`, температурой 0 и проверкой `finish_reason=stop`.
Ответ проверяется по схеме, цитата обязана буквально присутствовать в сегменте.
Относительный срок оставляется пустым до ручного подтверждения.

Riva не предоставляет совместимый `/audio/transcriptions`. Go запускает только
локальный `scripts/riva_transcribe.py`: существующий Riva Python SDK → gRPC.
Это вспомогательный процесс, не Python веб-сервер. Он декодирует MP3/WAV/M4A/MP4/
OGG/WebM/FLAC в PCM16 mono 16 kHz, отправляет последовательные фрагменты до 60 секунд,
сохраняет смещения timestamps и удаляет временный WAV. Максимум — 60 минут; границы
фрагментов могут снижать точность. Диаризации нет: все спикеры явно неизвестны.
KK вызывается только если сервер рекламирует kk-KZ; mixed пока отклоняется.

## Повторное использование окружения на VM

Самый быстрый путь без установки моделей: скопировать checkout Butterfly и собранный
Linux/amd64 Go бинарник в **уже работающий** контейнер
`multilingual-assistant-single-gpu`. Сеть и Riva SDK там уже использовались smoke.
На VM, после клонирования репозитория в `/home/ubuntu/butterfly`:

```bash
cd /home/ubuntu/nemotron-voice-agent
DC=(docker compose --env-file .env --profile multilingual-assistant/single-gpu --profile turn)
"${DC[@]}" cp /home/ubuntu/butterfly multilingual-assistant-single-gpu:/tmp/butterfly
"${DC[@]}" exec -T multilingual-assistant-single-gpu \
  uv run --frozen python -c 'import sys,riva.client; print(sys.executable)'
```

Запомните напечатанный абсолютный путь Python. Не копируйте `.env`, ключи и аудио в
репозиторий. В интерактивном терминале существующего контейнера:

```bash
cd /tmp/butterfly
export RIVA_PYTHON=/absolute/path/printed/above
export RIVA_ASR_ENDPOINT=nemo-speech-multilingual:50051
export ASR_LANGUAGE=ru-RU
export LOCAL_LLM_BASE_URL=http://nvidia-llm-vllm:8000/v1
export LOCAL_LLM_MODEL=nvidia/nemotron-3.5-lightning-30b-a3b
export BUTTERFLY_HOST=0.0.0.0
export PORT=8000
# Используйте отдельное Python-окружение с requirements.txt для PDF/DOCX,
# если reportlab/python-docx отсутствуют в существующем окружении.
export BUTTERFLY_EXPORT_PYTHON=/absolute/path/to/export-venv/bin/python
./bin/butterfly
```

Бинарник заранее собрать из этого checkout: `GOOS=linux GOARCH=amd64 go build -o
bin/butterfly ./cmd/butterfly`. Путь `bin/butterfly` не коммитится.
Файлы `/tmp` контейнера и процесс пропадут при пересоздании: для постоянного запуска
нужны отдельный контейнер/сервис и persistent volume. Эти команды не объявлены
проверенным deployment и не изменяют существующие model-контейнеры.

Альтернатива — процесс на хосте с собственным окружением Riva SDK и ffmpeg:
`RIVA_ASR_ENDPOINT=127.0.0.1:50051`,
`LOCAL_LLM_BASE_URL=http://127.0.0.1:18000/v1`. Все эти loopback-адреса относятся к VM,
не к локальному Codex. `NEMOTRON_BASE_URL/MODEL` также поддержаны и имеют приоритет
над `LOCAL_LLM_BASE_URL/MODEL`.

## Проверка и внешний доступ

Сначала `/health`, затем загрузка собственной записи через UI backend, проверка
реального транскрипта и цитат, подтверждение поручений и PDF/DOCX. Значение
`configured` означает наличие настроек, не успешную обработку на VM.

Brev Secure Links из handoff защищены владельцем и не являются публичным meeting API.
Текущий patch не создаёт публичный HTTPS endpoint, mapping нового порта или
авторизацию между Vercel и Brev. Для браузера Vercel нужны HTTPS backend, точный
CORS origin и cookies Secure/SameSite=None либо единый origin через reverse proxy.
Сырые порты Riva/LLM оставлять приватными. До браузерного upload/export smoke нельзя
объявлять frontend → Brev deployment готовым.
