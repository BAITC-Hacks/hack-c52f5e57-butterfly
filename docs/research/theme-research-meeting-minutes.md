# Theme research - butterfly-meeting-minutes

## Scope

Дата исследования: 2026-09-23. Тема: «Система автопротоколирования совещаний с фиксацией поручений». Заказчик результата — секретарь и руководитель совещания. Это исследовательский brief и передача в разработку, не отчёт о готовом продукте.

Подтверждено пользователем 2026-09-23: баланс NVIDIA Brev — 50 USD; GPU пользовательского компьютера не используем; демонстрация должна работать по внешней ссылке; frontend публикуем на Vercel в конце. Баланс принят со слов владельца, не проверен через аккаунт. Требуются интеграция Fly из HackAlem и Telegram. Срок сдачи пока неизвестен.

| № | Вопрос | Решение | Достаточное доказательство |
|---|---|---|---|
| Q1 | Какой сценарий обязателен и как оценивается? | Объём MVP и порядок работ | ТЗ, рубрика, воспроизводимый тест |
| Q2 | Как распознавать RU/KK/mixed и связывать речь с людьми? | ASR, диаризация, проверка имён | Официальные локальные реализации + тест на целевом GPU |
| Q3 | Как извлекать поручения и сроки без внешних AI API? | Self-hosted LLM и модель данных | Model cards, schema validation, эталонные примеры |
| Q4 | Как разместить Vercel + Brev в пределах бюджета? | Граница frontend/backend, эксплуатация | Документация провайдеров, доступный тариф и smoke test |
| Q5 | Как использовать «муху» с доказуемой пользой? | Переиспользование визуализации и feedback | Исходный код HackAlem, лицензии assets, связанный event trace |
| Q6 | Как добавить Telegram с учётом ТЗ? | Команды, уведомления, безопасные данные | Bot API, проверка доступа и доставка тестового события |

Confidence: high — прямой первичный источник или проверенный исходный код; medium — применимость подтверждена частично; low/unverified — гипотеза без runtime-проверки. Рекомендации ниже явно отделены от измеренных результатов.

## Questions and answers

### Q1. Обязательный продукт и рубрика — high

ТЗ требует speech-to-text RU, KK и смешанной речи; диаризацию с привязкой поручения к участнику; автоматическое выделение сути, ответственного и срока; итоговый протокол, summary и экспорт PDF/DOCX. В MVP планируем оба формата. Разрешены смоделированные записи; реальные записи для демонстрации должны быть обезличены. Внешние облачные API для аудио/текста запрещены, требуется переносимость в закрытый контур. [ТЗ организаторов](https://docs.google.com/document/d/1PUDYCg2OC_wfBk4rq5673yc1qsmo8O5_pwxjm483omU/edit?tab=t.0), 2026-09-23.

Рубрика именно этого кейса: работоспособность 25; техническая реализация 25; README/воспроизводимость 25; ценность 15; оригинальность 10. Это не стандартная рубрика Demo Day из плагина. Основной вклад в результат — рабочий сценарий и повторяемый запуск; Fly и Telegram добавляем после core. [ТЗ организаторов](https://docs.google.com/document/d/1PUDYCg2OC_wfBk4rq5673yc1qsmo8O5_pwxjm483omU/edit?tab=t.0), 2026-09-23.

Предлагаемый сценарий: загрузить постановочную встречу с согласившимися участниками → получить реплики с таймкодами → сопоставить Speaker 1/2 с именами → проверить поручения по цитатам → скачать протокол → увидеть событие завершения в Telegram и Fly. Подключение к Teams/Zoom/Meet упомянуто во входных данных, но отсутствует в отдельном списке обязательного минимума. Его обязательность требует уточнения у организаторов; upload-first — проектное допущение, не подтверждённое исключение.

### Q2. ASR и диаризация — high для возможностей; medium для выбора

Кандидат по умолчанию: self-hosted `faster-whisper`, multilingual `large-v3`, `task=transcribe`, затем `pyannote/speaker-diarization-community-1`. Whisper содержит RU/KK, faster-whisper умеет CPU/GPU, локальные weights и повторное определение языка по сегментам. Это не доказательство качества смешанной речи внутри одной фразы. На GPU Brev проведём отдельные RU/KK/mixed тесты до выбора окончательной конфигурации. [faster-whisper: официальная реализация](https://github.com/SYSTRAN/faster-whisper), [Whisper: языковые коды](https://github.com/openai/whisper/blob/main/whisper/tokenizer.py), [faster-whisper: transcribe.py](https://github.com/SYSTRAN/faster-whisper/blob/master/faster_whisper/transcribe.py), 2026-09-23.

Community-1 выдаёт интервалы/метки спикеров, обычную и exclusive diarization; имена автоматически не следуют из этих меток. Предлагается ручное сопоставление с roster через короткие аудиопримеры. Для первоначального скачивания нужны принятие условий HF и access token; затем предусмотрена загрузка pipeline с диска. Pipeline — CC-BY-4.0, код pyannote — MIT. [pyannote Community-1: model card](https://huggingface.co/pyannote/speaker-diarization-community-1), [pyannote.audio: runtime и telemetry](https://github.com/pyannote/pyannote-audio), 2026-09-23.

Offline-упаковка должна включать tokenizer/config/VAD, все веса и зависимости: в faster-whisper отсутствие tokenizer.json может вызвать отдельную загрузку даже при намерении работать локально. Отключить pyannote telemetry через `PYANNOTE_METRICS_ENABLED=0`; проверить холодный старт с запрещённым исходящим трафиком. [faster-whisper: transcribe.py](https://github.com/SYSTRAN/faster-whisper/blob/master/faster_whisper/transcribe.py), [pyannote.audio: runtime и telemetry](https://github.com/pyannote/pyannote-audio), 2026-09-23. Модели в этой исследовательской фазе не скачивались; latency, VRAM и точность не измерены.

### Q3. Поручения, сроки и summary — medium; качество unverified

Предлагаемый self-hosted кандидат: Qwen3-4B с отключённым reasoning или более крупная Qwen при подтверждённом запасе GPU; запуск через llama.cpp server, ограничение JSON-схемой и независимая валидация приложением. Официально заявлены multilingual возможности, включая казахский; Apache-2.0 указана в model card. Поддержка языка и структурного вывода не гарантирует фактической точности. [Qwen3-4B: официальная model card](https://huggingface.co/Qwen/Qwen3-4B), [Qwen3: языки](https://qwenlm.github.io/blog/qwen3/), [llama.cpp: server](https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md), 2026-09-23.

Проектная схема поручения: `action`, `speaker_id`, `assignee_id`, `deadline_raw`, `deadline_normalized`, `evidence_quote`, `segment_ids`, `start/end`, `needs_review`. Говорящий и ответственный различаются: «Иван, отправь отчёт» не поручение тому, кто произнёс фразу. Неопределённые имена/сроки оставляем null; относительные даты вычисляем относительно даты встречи и её timezone; исправления и отмены сохраняем в истории. Цитату проверяем по транскрипту, а имя — по подтверждённому roster. OpenAI/NVIDIA hosted inference не входят в обработку встречи из-за ограничения ТЗ.

### Q4. Размещение и бюджет — high для границ; medium для достаточности бюджета

Принятая архитектура по указанию пользователя: Vercel раздаёт веб-приложение; отдельный backend на GPU VM Brev принимает аудио напрямую по HTTPS, хранит задания и запускает self-hosted модели. На Vercel Functions лимит request/response payload составляет 4.5 MB; поэтому файлы не проксируем через функции. Асинхронный job API возвращает идентификатор; прогресс читается через polling/SSE backend. [Vercel Functions: ограничения](https://vercel.com/docs/functions/limitations), 2026-09-23. Это архитектурное решение, deployment ещё не выполнен.

Brev credits относятся к compute/storage, а не к эквивалентному балансу OpenAI или NVIDIA hosted API. 50 USD подтверждены пользователем. Карточки Launchables из сообщения — каталог шаблонов; итоговый выбранный GPU и тариф ещё не проверены. Расход зависит от instance, времени работы и диска; остановленный instance может сохранять оплачиваемый storage. [Brev Console: billing и compute](https://docs.nvidia.com/brev/guides/console-reference), 2026-09-23.

Предлагаемый бюджетный подход: одна GPU VM, последовательная загрузка моделей при необходимости, запуск на часы разработки/репетиции, контроль compute и storage. Выбор GPU и точный лимит часов фиксируются после просмотра доступного тарифа; достаточность 50 USD сейчас не доказана. В исследовании аренда не запускалась.

Публичное Brev/Vercel демо используем со смоделированными встречами. Для закрытого контура переносим frontend, API, модели и хранилище на инфраструктуру заказчика, внешние интеграции отключаем. Self-hosted модель на арендованной VM сама по себе не доказывает соответствие закрытому контуру — требуется отдельный тест переносимости и, при строгой трактовке, разъяснение организаторов. [ТЗ организаторов](https://docs.google.com/document/d/1PUDYCg2OC_wfBk4rq5673yc1qsmo8O5_pwxjm483omU/edit?tab=t.0), 2026-09-23.

### Q5. Fly — high для существующих компонентов; польза unverified

Проверенный код HackAlem имеет `/brain-demo`, `/fly-live` SSE/snapshot, физическую визуализацию тела и реакцию на recommend/feedback из журнала. Ручные reward/pain — симуляция. Fly не выполняет ASR или диаризацию. Runtime assets в `data/fly-viz` отсутствуют, текущий HTML использует Three.js CDN; для воспроизводимости нужны локальные assets и зависимости. Источник: HackAlem `cmd/demo-api/main.go`, `static/brain-demo.html`, `static/brain-demo.js`, `static/fly-anatomy.js`, `static/fly-brain.js`, проверено 2026-09-23.

Тело можно восстановить из DeepMind Menagerie `flybody`, Apache-2.0. Две геометрии мозга соответствуют medulla/saddle из VFB JRC2018; это области, не полный коннектом. Карточки этих данных указывают CC-BY-NC-SA-4.0; сохраняем attribution и не переносим лицензию MIT сайта на meshes. Для коммерческого варианта потребуется оценить совместимость либо заменить эти необязательные геометрии. [DeepMind Menagerie: flybody](https://github.com/google-deepmind/mujoco_menagerie/tree/main/flybody), [Flybody: Apache-2.0](https://github.com/google-deepmind/mujoco_menagerie/blob/main/flybody/LICENSE), [VFB: medulla JRC2018](https://www.virtualflybrain.org/term/me-on-jrc2018unisex-adult-brain-vfb_00102107/), [VFB: saddle JRC2018](https://www.virtualflybrain.org/term/sad-on-jrc2018unisex-adult-brain-vfb_00102271/), 2026-09-23.

Предложение для продукта: Fly показывает реальные переходы job и результаты проверки поручений; recommend/feedback работает как экспериментальная рекомендация следующего действия (проверить владельца, срок, цитату). Feedback записывается один раз после реального подтверждения/исправления, связан с event ID. Если адаптация маршрута не реализована и не измерена, UI называется визуализацией, а не обучением мозга. Отказ Fly не должен останавливать протоколирование. Пользу проверяем сравнением с обычной очередью проверки на одинаковом наборе встреч.

### Q6. Telegram — high для протокола; интеграция unverified

Предлагаемые команды: `/start`, `/help`, `/status`, `/subscribe`, `/unsubscribe`; уведомления «обработка завершена/ошибка». BotFather создаёт бота; пользователь должен первым начать диалог. Проверять chat/user allowlist на сервере; command scope не является авторизацией. Для backend Brev достаточно long polling отдельным worker; альтернатива — HTTPS webhook с secret header. Polling и webhook взаимоисключающие. [Telegram Bot API](https://core.telegram.org/bots/api), [Telegram: создание бота и команды](https://core.telegram.org/bots/features), 2026-09-23.

Для повторных update нужна дедупликация; для rate limits — очередь и retry_after. Токен только в server-side secret, не в frontend и логах. Telegram — внешний сервис: в режиме закрытого контура выключен, для публичного смоделированного демо включается явно; в уведомлении по умолчанию нет транскриптов, имён и содержания поручений. Это проектное решение из ограничения ТЗ. [Telegram Bot API](https://core.telegram.org/bots/api), [Telegram: ограничения отправки](https://core.telegram.org/bots/faq#my-bot-is-hitting-limits-how-do-i-avoid-this), [ТЗ организаторов](https://docs.google.com/document/d/1PUDYCg2OC_wfBk4rq5673yc1qsmo8O5_pwxjm483omU/edit?tab=t.0), 2026-09-23. Бот/token/chat пока не настроены; тестовых отправок не было.

## Evidence table

| Claim | Source | Date | Confidence | Confirmations | Notes |
|---|---|---|---|---|---|
| Обязательные функции, внешние API запрещены; синтетическое демо разрешено | [ТЗ организаторов](https://docs.google.com/document/d/1PUDYCg2OC_wfBk4rq5673yc1qsmo8O5_pwxjm483omU/edit?tab=t.0) | 2026-09-23 | high | — | Corpus chunks 2–3; возможность переноса ещё не проверена |
| Рубрика 25/25/25/15/10 | [ТЗ организаторов](https://docs.google.com/document/d/1PUDYCg2OC_wfBk4rq5673yc1qsmo8O5_pwxjm483omU/edit?tab=t.0) | 2026-09-23 | high | — | Corpus chunk 4; использовать вместо общего шаблона плагина |
| Локальный faster-whisper поддерживает CPU/GPU и multilingual модели | [faster-whisper: официальная реализация](https://github.com/SYSTRAN/faster-whisper) | 2026-09-23 | high | — | Производительность целевой VM не измерена |
| В Whisper присутствуют RU/KK | [Whisper: языковые коды](https://github.com/openai/whisper/blob/main/whisper/tokenizer.py) | 2026-09-23 | high | — | Не доказывает качество mixed |
| multilingual по сегментам; нужен комплект tokenizer | [faster-whisper: transcribe.py](https://github.com/SYSTRAN/faster-whisper/blob/master/faster_whisper/transcribe.py) | 2026-09-23 | high | — | Offline cold start обязателен |
| Community-1: offline, exclusive diarization, gated download, CC-BY-4.0 | [pyannote Community-1: model card](https://huggingface.co/pyannote/speaker-diarization-community-1) | 2026-09-23 | high | — | Метки спикеров не являются именами |
| pyannote telemetry можно отключить | [pyannote.audio: runtime и telemetry](https://github.com/pyannote/pyannote-audio) | 2026-09-23 | high | — | Отключение нужно проверить в deployment |
| Whisper допускает галлюцинации и различия качества языков | [Whisper large-v3 model card](https://huggingface.co/openai/whisper-large-v3) | 2026-09-23 | high | — | Нужны собственные тесты |
| Qwen3-4B имеет Apache-2.0 и self-hosted пути запуска | [Qwen3-4B: официальная model card](https://huggingface.co/Qwen/Qwen3-4B) | 2026-09-23 | high | — | Качество поручений пока unverified |
| Qwen3 заявляет казахский среди поддерживаемых языков | [Qwen3: языки](https://qwenlm.github.io/blog/qwen3/) | 2026-09-23 | high | — | Применимость к смешанным встречам medium |
| llama.cpp server поддерживает структурный вывод | [llama.cpp: server](https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md) | 2026-09-23 | high | — | JSON-схема не проверяет факты |
| Brev billing относится к compute/storage | [Brev Console: billing и compute](https://docs.nvidia.com/brev/guides/console-reference) | 2026-09-23 | high | — | 50 USD — отдельное подтверждение пользователя |
| Vercel Functions payload ограничен 4.5 MB | [Vercel Functions: ограничения](https://vercel.com/docs/functions/limitations) | 2026-09-23 | high | — | Файлы загружаются напрямую в backend |
| Telegram: polling/webhook, secret header, offset и sendMessage | [Telegram Bot API](https://core.telegram.org/bots/api) | 2026-09-23 | high | — | Runtime пока отсутствует |
| Flybody Menagerie — Apache-2.0 | [Flybody: Apache-2.0](https://github.com/google-deepmind/mujoco_menagerie/blob/main/flybody/LICENSE) | 2026-09-23 | high | — | Сохранить происхождение assets |
| Medulla ROI JRC2018 — CC-BY-NC-SA-4.0 | [VFB: medulla JRC2018](https://www.virtualflybrain.org/term/me-on-jrc2018unisex-adult-brain-vfb_00102107/) | 2026-09-23 | high | — | Не полный мозг |
| Saddle ROI JRC2018 — CC-BY-NC-SA-4.0 | [VFB: saddle JRC2018](https://www.virtualflybrain.org/term/sad-on-jrc2018unisex-adult-brain-vfb_00102271/) | 2026-09-23 | high | — | Нет checksum прежних потерянных копий |
| Баланс Brev 50 USD; Vercel в конце; не использовать личный GPU | Прямые сообщения пользователя в этой задаче | 2026-09-23 | high для требований владельца | — | Billing недоступен без входа; тариф не выбран |
| Fly demo читает события и зависит от отсутствующих assets | HackAlem cmd/demo-api и static/* | 2026-09-23 | high | — | Проверка исходников, не работающего сайта |

Coverage: 2 из 6 вопросов закрыты на уровне требований и официального протокола (Q1, Q6); 4 частично закрыты (Q2–Q5), runtime-качество, цена выбранной VM, публикация и эффект Fly остаются открытыми. Обязательность meeting-bot из Q1 вынесена как уточнение. Все 6 вопросов имеют ответы; ни одна интеграция не объявлена завершённой. Web-доступ использован. Google Docs прочитан через текстовый export после неудачного web-open. Brev Billing требует авторизации. Это не блокирует сборку кода, но блокирует подтверждение реального GPU deployment.

Corpus evidence: `docs/research/corpus-evidence.json`; extracted artifact `extracted/meeting-minutes-spec-37aeaa58b721f206-20260923-113354.138606678.txt`; index `knowledge/meeting-minutes-spec-37aeaa58b721f206-20260923-113354-138606-5035bc731e944458.json`. Исходный SHA256 указан в `source-metadata.json`.

## Implications for the build

1. Собрать переносимый monorepo: frontend для Vercel; API и worker для Brev; отдельные profiles публичного synthetic-demo и offline/on-premise. Веса не включать в frontend или git; manifest фиксирует revision, checksum и лицензии.
2. API: создать встречу с roster/date/timezone; принять ограниченный по размеру/длительности файл; вернуть job ID; отдать прогресс, transcript, reviewable action items и локально сформированные PDF/DOCX. HTTPS, авторизация на job, строгий CORS для выбранного frontend, лимит очереди и таймауты. Прямой upload не означает открытый GPU endpoint без защиты.
3. Постановочная встреча — тот же inference path, что и пользовательская загрузка. Предрассчитанный пример допускается только с заметной маркировкой; при недоступном GPU честный статус ошибки, без выдачи фикстуры за анализ нового файла.
4. Сначала вертикальный сценарий, затем Fly/Telegram. Состояния: uploaded, queued, transcribing, diarizing, extracting, needs_review, completed, failed. События имеют ID; текст встречи не копируется в публичные логи визуализации.
5. Проверка: 8 коротких постановочных клипов — RU, KK, переключение между репликами, mixed внутри фразы, третий участник как ответственный, перекрытие, тишина/шум, отмена/перенос. Один длинный прогон около 10 минут. Это план тестирования, не выполненный benchmark.
6. Измерять WER/CER по языкам, ошибки speaker attribution, точность action/owner/deadline, число необоснованных поручений, время и VRAM. Обязательный gate: не выдавать неопределённого владельца или срок за подтверждённые; сохранять цитату/таймкод.
7. README должен позволять повторить запуск, загрузку весов и один обязательный сценарий. Offline cold start проверяет отсутствие обращения к внешнему inference/CDN/telemetry. Публикация Vercel — в финале после работающего backend и E2E проверки.

## Risks and unknowns

- Срок сдачи неизвестен: оценки времени в плане предварительные, не обещание уложиться в неизвестный дедлайн.
- Доступ к Brev в доступной browser-session отсутствует; VM, точный GPU/тариф, HTTPS endpoint и часы аренды ещё не определены. 50 USD — баланс со слов пользователя; это не доказательство достаточности или разрешение автоматически потратить всю сумму.
- Vercel project/deployment ещё не создан. Требуется доступ к выбранному аккаунту на этапе публикации; код до этого можно собирать и проверять.
- Модели/веса и HF gated access Community-1 не подготовлены. Не измерены RU/KK/mixed качество, latency и VRAM.
- Название Whisper имеет различающиеся license metadata: официальный OpenAI repo сообщает MIT, HF OpenAI card — Apache-2.0; сохранять лицензию конкретного скачанного артефакта и provenance Systran conversion, не заменять её догадкой.
- OpenAI API не входит в core: аудио и текст обрабатываются self-hosted моделями согласно ограничениям ТЗ.
- В Fly отсутствуют runtime meshes; повторяемое получение и оптимизация загрузки ещё не реализованы. До измерений не заявлять повышение качества или обучение ASR от Fly.
- Telegram bot/chat и server-side token отсутствуют; отправка не проверялась. Для чувствительных данных Telegram исключён из offline profile.
- Требуются уточнения организаторов: обязательность подключения участником к Teams/Zoom/Meet и достаточность публичного синтетического демо вместе с переносимым self-hosted стеком для on-premise ограничения.

## Sources

- [ТЗ организаторов](https://docs.google.com/document/d/1PUDYCg2OC_wfBk4rq5673yc1qsmo8O5_pwxjm483omU/edit?tab=t.0) — retrieved 2026-09-23.
- [faster-whisper: официальная реализация](https://github.com/SYSTRAN/faster-whisper) — retrieved 2026-09-23.
- [faster-whisper: transcribe.py](https://github.com/SYSTRAN/faster-whisper/blob/master/faster_whisper/transcribe.py) — retrieved 2026-09-23.
- [Whisper large-v3 model card](https://huggingface.co/openai/whisper-large-v3) — retrieved 2026-09-23.
- [Whisper: языковые коды](https://github.com/openai/whisper/blob/main/whisper/tokenizer.py) — retrieved 2026-09-23.
- [Systran: конвертированные large-v3 weights](https://huggingface.co/Systran/faster-whisper-large-v3) — retrieved 2026-09-23.
- [pyannote Community-1: model card](https://huggingface.co/pyannote/speaker-diarization-community-1) — retrieved 2026-09-23.
- [pyannote.audio: runtime и telemetry](https://github.com/pyannote/pyannote-audio) — retrieved 2026-09-23.
- [Qwen3-4B: официальная model card](https://huggingface.co/Qwen/Qwen3-4B) — retrieved 2026-09-23.
- [Qwen3: языки](https://qwenlm.github.io/blog/qwen3/) — retrieved 2026-09-23.
- [llama.cpp: server](https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md) — retrieved 2026-09-23.
- [Brev Console: billing и compute](https://docs.nvidia.com/brev/guides/console-reference) — retrieved 2026-09-23.
- [Brev: самостоятельное развёртывание NIM](https://docs.nvidia.com/brev/guides/inference-deployment/deploying-nims) — retrieved 2026-09-23.
- [Vercel Functions: ограничения](https://vercel.com/docs/functions/limitations) — retrieved 2026-09-23.
- [Telegram Bot API](https://core.telegram.org/bots/api) — retrieved 2026-09-23.
- [Telegram: создание бота и команды](https://core.telegram.org/bots/features) — retrieved 2026-09-23.
- [Telegram: ограничения отправки](https://core.telegram.org/bots/faq#my-bot-is-hitting-limits-how-do-i-avoid-this) — retrieved 2026-09-23.
- [DeepMind Menagerie: flybody](https://github.com/google-deepmind/mujoco_menagerie/tree/main/flybody) — retrieved 2026-09-23.
- [Flybody: Apache-2.0](https://github.com/google-deepmind/mujoco_menagerie/blob/main/flybody/LICENSE) — retrieved 2026-09-23.
- [VFB: medulla JRC2018](https://www.virtualflybrain.org/term/me-on-jrc2018unisex-adult-brain-vfb_00102107/) — retrieved 2026-09-23.
- [VFB: saddle JRC2018](https://www.virtualflybrain.org/term/sad-on-jrc2018unisex-adult-brain-vfb_00102271/) — retrieved 2026-09-23.
- Локальный первичный источник: HackAlem 0.16.66, `cmd/demo-api/main.go`, `cmd/demo-api/static/*`, skills `theme-research`, `fly-protocol`, `telegram-integration` — inspected 2026-09-23.
- Требования владельца: сообщения текущей задачи о Brev/Vercel и запрете личного GPU — 2026-09-23.

## What would change our mind

- Q1: организаторы подтвердят обязательный live meeting-bot — добавляем коннектор в core, сокращаем необязательный Fly-полиш.
- Q2: large-v3 не справляется с KK/mixed, временем или VRAM на тестовых клипах — сравниваем меньшую multilingual модель либо документированный KK/RU fine-tune; меняем выбор по собственным данным.
- Q3: Qwen пропускает поручения/галлюцинирует имена и сроки — усиливаем review/валидацию, сравниваем другой self-hosted размер; облачный API не становится автоматическим fallback.
- Q4: доступный тариф не укладывается в бюджет или организаторы запрещают даже синтетическое облачное демо — пересматриваем аренду/режим показа. Личный GPU исключён указанием пользователя. Vercel остаётся финальным frontend, пока пользователь не изменит это решение.
- Q5: Fly не улучшает проверку или ухудшает стабильность — оставляем явно обозначенную визуализацию либо сокращаем эксперимент; не заявляем подтверждённое преимущество.
- Q6: политика заказчика запрещает любые внешние уведомления — полностью отключаем Telegram; основной сценарий продолжает работать.
