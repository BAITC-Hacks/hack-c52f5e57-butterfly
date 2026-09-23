# Nemotron в Brev: runtime и первый сквозной тест

**Локальные ASR и LLM работают на выбранной GPU VM.** Первые 30 секунд второй
тестовой встречи прошли цепочку «аудио → русский текст → JSON поручений».
Этот отчёт заменяет прежний статус «подготовка / inference не проверен» из
[отчёта подготовки](nemotron-preparation-report.md); исторический envelope сохранён.
Машиночитаемая передача контекста: [runtime-handoff.json](runtime-handoff.json).

## Подтверждённое состояние

Проверка health выполнена **2026-09-23 12:32:53 UTC** на VM
`nemotron-voice-agent-402b8f`: NVIDIA RTX PRO 6000 Blackwell Server Edition,
97887 MiB VRAM, 8 vCPU, 64 GiB RAM, 500 GiB диска. Отображённая цена VM —
$4.12/час. Профиль — `multilingual-assistant/single-gpu` вместе с `turn`;
upstream commit — `13801d9963aea20ac663318060824a78b28809fa`.

Четыре контейнера запущены; приложение и vLLM имеют состояние healthy.

| Сервис | Проверка | Результат |
|---|---|---|
| Приложение | `127.0.0.1:7860/health` | HTTP 200 |
| vLLM | `127.0.0.1:18000/health` | HTTP 200 |
| vLLM | `127.0.0.1:18000/v1/models` | `nvidia/nemotron-3.5-lightning-30b-a3b` |
| ASR | Riva gRPC, `127.0.0.1:50051` | TCP доступен; Recognize прошёл тест |

[Voice UI](https://nemotron-voice-agent-vt4eo8rll.gobrev.dev/) инициализирован,
доступ требует входа владельца в Brev. В интерфейсе доступны `de-DE`, `en-US`,
`es-ES`, `fr-FR`, `it-IT`. Русский ASR проверен через API; русская voice-сессия
в штатном интерфейсе этим не подтверждена. Публичные endpoints сырых ASR/LLM
дополнительно не открывались.

## Результат smoke test

Клип преобразован в mono 16 kHz PCM16 и передан в локальный Riva Recognize с
`ru-RU`. Затем текст обработан локальным Nemotron через `/v1/chat/completions`.

| Метрика | Результат |
|---|---:|
| Длительность клипа | 30.0 с |
| Время ASR | 0.930 с |
| ASR real-time factor | 0.031 |
| Размер полученного текста | 455 символов |
| Время LLM | 2.033 с |
| Поручений в JSON | 1 |

Проверены структура JSON и наличие цитат поручения в полученном тексте.
Это успешный технический тест короткого фрагмента: WER, качество поручений,
правильность ответственного и срока, а также диаризация **не оценивались**.
Эти измерения не являются временем полной обработки встречи и не включают
загрузку моделей или развёртывание.

Запуск завершился с **exit code 0** через Jupyter subprocess runner с `check=True`.
Точная выполненная команда, working directory `/home/ubuntu/nemotron-voice-agent`:

```sh
docker compose --profile multilingual-assistant/single-gpu --profile turn exec -T multilingual-assistant-single-gpu uv run --frozen python /tmp/smoke.py /tmp/meeting-2.mp3 --container-network --out /tmp/meeting2-smoke-1
```

Ячейка notebook 18 завершилась с execution count `[12]`, kernel перешёл в idle;
исключений не было. Для повторного запуска нужно указать новый output directory.
Решающий результат проверки:

```json
{"status":"passed","out":"/tmp/meeting2-smoke-1","audio_seconds":30.0,"asr_seconds":0.9295554129998891,"asr_real_time_factor":0.030985180433329637,"transcript_characters":455,"llm_seconds":2.03318032199968,"actions_count":1,"accuracy_evaluated":false,"diarization_evaluated":false}
```

Полные результаты подтверждённо сохранены на VM в
`/home/ubuntu/butterfly-test-results/meeting2-smoke-1`: каталог `0700`, файлы `0600`.
Соседний `meeting2-smoke-1-manifest.json` также имеет права `0600`.
В handoff записаны размеры и SHA256 всех пяти файлов: `clip.wav`, `clip.json`,
`asr.json`, `actions.json`, `result.json`. SHA256 клипа:
`4c00047efd5d9f2e2d0606e2a925e8a332b39dd9ea2f5b04f02744c1c2050776`.
Исходный output directory контейнера — `/tmp/meeting2-smoke-1`; notebook сохранён.
Аудио, стенограммы и содержимое поручений в отчёт и публичный Git не включены.
SHA256 загруженного клиента проверен перед `docker compose cp`:
`fe08a1e9c37cc7611acc2a97ad9f492e265fbc21f7c9999605c3b55b83d6693d`.
Отдельный hash файла внутри контейнера не снимался; это различие сохранено в handoff.

## Подключение будущего meeting adapter

Ниже предложены переменные для нашего адаптера; это не новые параметры upstream.

| Переменная | Процесс на хосте VM | Контейнер в той же Compose-сети |
|---|---|---|
| `RIVA_ASR_ENDPOINT` | `127.0.0.1:50051` | `nemo-speech-multilingual:50051` |
| `LOCAL_LLM_BASE_URL` | `http://127.0.0.1:18000/v1` | `http://nvidia-llm-vllm:8000/v1` |
| `LOCAL_LLM_MODEL` | `nvidia/nemotron-3.5-lightning-30b-a3b` | то же |
| `ASR_LANGUAGE` | `ru-RU` | то же |

Speech и LLM в тесте использовали локальную Compose-сеть. Проверка окружения
приложения **2026-09-23 12:36:29 UTC** подтвердила
`external_nvidia_api_key_set=false`; host `.env` имеет права `0600`.
`NVIDIA_API_KEY` очищен, ключ для облачного inference fallback отсутствует.
Проверка полностью автономной работы без внешней сети после загрузки образов
и весов ещё не выполнена.

Далее нужны: приём записей и очередь задач; полный прогон двух встреч по эталонам;
диаризация для пяти спикеров; отдельный путь KK/mixed; проверка поручений и сроков;
PDF/DOCX; Fly и Telegram; финальный frontend на Vercel. Готовность этих функций
текущим smoke test не заявляется.

Исходники проверенного стека: [Blueprint и закреплённый commit](https://github.com/NVIDIA-AI-Blueprints/nemotron-voice-agent/tree/13801d9963aea20ac663318060824a78b28809fa),
[gRPC ASR runtime v0.1.0](https://github.com/NVIDIA/NeMo-Speech.cpp/blob/v0.1.0/src/services/grpc_asr.cc).
