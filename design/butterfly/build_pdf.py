"""Build the Butterfly visual-design handoff. Requires reportlab."""
from pathlib import Path
from reportlab.pdfgen import canvas
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.lib.colors import HexColor
from reportlab.lib.utils import ImageReader

ROOT = Path(__file__).resolve().parent
OUT = ROOT / 'output/pdf/butterfly-website-design.pdf'
FONT_ROOT = Path('/usr/share/fonts/truetype/noto')
pdfmetrics.registerFont(TTFont('Noto', str(FONT_ROOT / 'NotoSans-Regular.ttf')))
pdfmetrics.registerFont(TTFont('NotoBold', str(FONT_ROOT / 'NotoSans-Bold.ttf')))
W, H = 1600, 1200
C = dict(bg='#F7F6F2', paper='#FFFFFF', text='#20231F', muted='#686D67',
         line='#E5E3DB', orange='#BD482E', pale='#F8EDE5', sage='#EAF0E9',
         green='#246B45', amber='#9B4B18')

OUT.parent.mkdir(parents=True, exist_ok=True)
pdf = canvas.Canvas(str(OUT), pagesize=(W, H), pageCompression=1)
pdf.setTitle('Butterfly - дизайн сайта и передача в разработку')
pdf.setAuthor('Butterfly / Design concept')
pdf.setSubject('Визуальный макет. Синтетические данные. Nemotron: целевой провайдер.')

def box(x, y, w, h, color='paper', radius=14, stroke=None):
    pdf.setFillColor(HexColor(C.get(color, color)))
    pdf.setStrokeColor(HexColor(C.get(stroke, stroke) if stroke else C.get(color, color)))
    pdf.setLineWidth(1)
    pdf.roundRect(x, y, w, h, radius, fill=1, stroke=bool(stroke))

def text(x, y, value, size=18, color='text', bold=False):
    pdf.setFillColor(HexColor(C.get(color, color)))
    pdf.setFont('NotoBold' if bold else 'Noto', size)
    pdf.drawString(x, y, value)

def lines(x, y, values, size=18, gap=29, color='muted', bold=False):
    for line in values:
        text(x, y, line, size, color, bold)
        y -= gap
    return y

def rule(x, y, w):
    pdf.setStrokeColor(HexColor(C['line']))
    pdf.setLineWidth(1)
    pdf.line(x, y, x + w, y)

def badge(x, y, label, color='sage', ink='green', width=None):
    width = width or pdfmetrics.stringWidth(label, 'Noto', 15) + 28
    box(x, y, width, 32, color, 8)
    text(x + 14, y + 10, label, 15, ink)

def header(page, title, subtitle):
    box(0, 0, W, H, 'bg', 0)
    text(48, 1155, 'BUTTERFLY  /  DESIGN HANDOFF', 15, 'orange', True)
    text(1450, 1155, f'0{page} / 03', 15, 'muted')
    text(48, 1110, title, 30, bold=True)
    text(48, 1081, subtitle, 16, 'muted')

def footer(page):
    rule(48, 45, 1504)
    text(48, 22, 'Визуальный концепт v1.0  •  23.09.2026  •  Все данные вымышлены', 13, 'muted')
    text(1180, 22, 'Nemotron: целевой провайдер', 13, 'muted')
    pdf.showPage()

def screen(page, title, subtitle, filename):
    header(page, title, subtitle)
    img = ImageReader(str(ROOT / 'assets' / filename))
    iw, ih = img.getSize()
    # Original aspect ratio, no clipping, no resampling or image editing.
    scale = min(1504 / iw, 1000 / ih)
    width, height = iw * scale, ih * scale
    pdf.drawImage(img, 48 + (1504 - width) / 2, 64 + (1000 - height) / 2,
                  width=width, height=height, preserveAspectRatio=True, mask='auto')
    footer(page)

screen(1, 'Проверка поручений',
       'Главный рабочий экран: источник, ответственный, срок и подтверждение человеком.',
       '01-meeting-review.png')
screen(2, 'Встречи и новая запись',
       'Вход в сценарий: загрузка записи, поиск встреч и понятный статус обработки.',
       '02-meetings-upload.png')

header(3, 'Дизайн-система и правила реализации',
       'Эта страница определяет точные токены и поведение; изображения задают визуальное направление.')

# Three aligned columns: foundations, components, behavior.
for x in (48, 560, 1072):
    box(x, 440, 480, 604, 'paper', 14, 'line')

text(76, 1006, '01  Основа', 23, bold=True)
palette = [('Фон', 'bg'), ('Поверхность', 'paper'), ('Текст', 'text'),
           ('Акцент', 'orange'), ('Граница', 'line'), ('Подтверждено', 'green')]
for i, (label, token) in enumerate(palette):
    col, row = i % 3, i // 3
    x, y = 76 + col * 143, 883 - row * 114
    box(x, y + 35, 124, 48, token, 8, 'line')
    text(x, y + 12, label, 15, 'text', True)
    text(x, y - 10, C[token].upper(), 13, 'muted')
rule(76, 740, 424)
text(76, 704, 'Типографика', 19, bold=True)
lines(76, 674, ['Manrope; fallback Noto Sans.', 'Заголовок 32/40, semibold.',
                 'Раздел 20/28; текст 16/24.', 'Метки и метаданные 13/20.'], 17, 29)
text(76, 535, 'Сетка и ритм', 19, bold=True)
lines(76, 507, ['Шаг 8 px; поля 32 px; gap 24 px.', 'Радиусы 12 px; границы 1 px.'], 17, 29)

text(588, 1006, '02  Компоненты', 23, bold=True)
box(588, 924, 212, 50, 'orange', 10)
text(612, 941, 'Подтвердить', 18, 'paper', True)
box(816, 924, 192, 50, 'paper', 10, 'line')
text(846, 941, 'Экспорт', 18, 'text', True)
badge(588, 868, 'Подтверждено')
badge(777, 868, 'Нужна проверка', 'pale', 'amber')
text(588, 819, 'Ответственный', 15, 'muted')
box(588, 758, 420, 48, 'paper', 8, 'orange')
text(604, 774, 'Назначить участника', 17, 'muted')
text(588, 725, 'Не определён в записи', 14, 'amber')
rule(588, 699, 420)
lines(588, 665, ['Карточка поручения:', 'действие, владелец, дата, статус,', 'цитата и переход по таймкоду.'], 17, 29)
lines(588, 553, ['Минимум 44 px для нажатия.', 'Focus: обводка 2 px + отступ.', 'Статус всегда содержит текст.'], 17, 29)

text(1100, 1006, '03  Состояния и адаптация', 23, bold=True)
text(1100, 957, 'Последовательность', 18, bold=True)
lines(1100, 926, ['Загрузка / Распознавание', 'Извлечение / Проверка / Готово'], 17, 29)
rule(1100, 869, 424)
text(1100, 835, 'Отдельные состояния', 18, bold=True)
lines(1100, 804, ['Пустой список; файл выбран;', 'обработка; ошибка и повтор;', 'провайдер недоступен.'], 17, 29)
rule(1100, 711, 424)
text(1100, 677, 'Desktop / tablet / mobile', 18, bold=True)
lines(1100, 646, ['От 1280: sidebar 224, контекст 288.', '768-1279: sidebar 72, контекст ниже.', 'До 768: один столбец, меню скрыто;', 'таблица превращается в карточки.', 'Поля и кнопки без горизонтального', 'скролла; аудио-панель в потоке.'], 16, 29)

box(48, 144, 1504, 260, 'sage', 14)
text(76, 363, 'Передача в активную разработку', 23, bold=True)
lines(76, 321, ['NVIDIA Nemotron - основной планируемый провайдер. UI получает статус от backend.',
                'Модель, endpoint, прогресс и качество RU/KK не выдумываются и не зашиваются в frontend.',
                'Неизвестные владелец и срок остаются пустыми до подтверждения. Цитата ведёт к источнику.'], 19, 35, 'text')
text(76, 221, 'Brev-демо: «Обработка на сервере проекта»; обещание закрытого контура - только для on-premise.', 16, 'muted')
lines(76, 193, ['PDF/DOCX доступны после проверки. Fly и Telegram подключаются к реальным событиям backend.',
                'Макеты не подтверждают работающий inference. Изображение статуса не является health-check.'], 17, 29, 'muted')
text(48, 94, 'Передавать только design/butterfly/. Код приложения и общий план в другой ветке не меняются.', 17, 'muted')
footer(3)
pdf.save()
print(OUT)
