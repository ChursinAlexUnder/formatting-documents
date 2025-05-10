import re
from docx.oxml import OxmlElement
from docx.oxml.ns import qn
from collections import Counter
from docx.shared import Pt, Cm, RGBColor
from docx.enum.text import WD_PARAGRAPH_ALIGNMENT
from docx.enum.style import WD_STYLE_TYPE

def headingLevel(text):
    """ Определяет уровень заголовка."""
    text_lower = text.lower()
    if (text_lower in ("содержание", "введение", "заключение", "реферат", "приложение")
        or (text_lower.startswith("список") and ("источников" in text_lower or "литературы" in text_lower))):
        return 1
    match = re.match(r"^\d+(\.\d+){0,3}", text)
    return min(len(match.group().split(".")), 4) if match else False

def isHeading(paragraph):
    """ Проверяет, является ли абзац заголовком. """
    text = paragraph.text.strip()
    text_lower = text.lower()
    
    # Проверка на пустой текст или слишком длинный абзац
    if not text or len(text) > 150:
        return False
    
    if (text_lower in ("содержание", "введение", "заключение", "реферат", "приложение")
        or (text_lower.startswith("список") and ("источников" in text_lower or "литературы" in text_lower))):
        return True

    # Проверка, если абзац содержит только картинку/формулу
    if not paragraph.runs:  # Если нет run'ов, значит, текст отсутствует (например, изображение)
        return False

    bold_count = 0
    total_chars = 0

    # Получаем стиль параграфа (если есть)
    paragraph_style = paragraph.style
    style_is_bold = False

    if paragraph_style and paragraph_style.font:
        style_is_bold = paragraph_style.font.bold

    # Проходимся по всем run'ам в абзаце
    for run in paragraph.runs:
        run_text = run.text.strip()
        if not run_text:  # Пропускаем пустые run'ы
            continue

        total_chars += len(run_text)

        # Определяем жирность с учётом стиля параграфа
        run_is_bold = run.bold if run.bold is not None else style_is_bold
        if run_is_bold:
            bold_count += len(run_text)

    # Если нет текста (например, все run'ы пустые), то это не заголовок
    if total_chars == 0:
        return False

    # Проверяем процентное соотношение
    bold_percentage = (bold_count / total_chars) * 100

    # Условия определения заголовка
    if (bold_percentage >= 90 or total_chars - bold_count <= 3):
        return True
    return False

def removeEmptyLinesAndPageBreaks(doc, index):
    """Удаляет пустые строки, пустые абзацы и разрывы страниц перед и после параграфа, не затрагивая контент (картинки, таблицы)."""
    ns = {'w': 'http://schemas.openxmlformats.org/wordprocessingml/2006/main'}

    def is_empty_paragraph(paragraph):
        """
        Проверяет, является ли абзац пустым.
        Абзац считается непустым, если:
          - содержит текст (даже если `paragraph.text` пуст),
          - содержит встроенные объекты (например, картинки).
        """
        # Если есть текст, отличающийся от пробелов, считаем абзац не пустым.
        if paragraph.text and paragraph.text.strip():
            return False

        if paragraph._element.find('.//w:drawing', namespaces=ns) is not None or paragraph._element.find('.//w:pict', namespaces=ns) is not None:
            return False  # В абзаце есть объект (например, картинка)

        return True  # Если ничего нет, считаем абзац пустым

    def remove_paragraph(paragraph):
        """Удаляет абзац из документа."""
        parent = paragraph._element.getparent()
        if parent is not None:
            parent.remove(paragraph._element)

    def remove_page_breaks(paragraph):
        """Удаляет только разрывы страниц внутри параграфа, оставляя остальной текст нетронутым."""
        for br in paragraph._element.findall(".//w:br[@w:type='page']", namespaces=ns):
            br.getparent().remove(br)

     # Удаление пустых абзацев и разрывов страниц перед указанным индексом
    i = index - 1
    while i >= 0:
        if i >= len(doc.paragraphs):
            break

        paragraph = doc.paragraphs[i]
        page_break = paragraph._element.find(".//w:br[@w:type='page']", namespaces=ns)

        if page_break is not None:
            remove_page_breaks(paragraph)  # Удаляем разрыв страницы, но не весь параграф
        elif is_empty_paragraph(paragraph):
            remove_paragraph(paragraph)
            index -= 1  # Так как удаляется абзац, индекс смещается
        else:
            break
        i -= 1

    # Удаляем пустые абзацы и разрывы страниц после целевого абзаца
    removed_after = 0
    i = index + 1
    while i < len(doc.paragraphs):
        paragraph = doc.paragraphs[i]
        page_break = paragraph._element.find(".//w:br[@w:type='page']", namespaces=ns)

        if page_break is not None:
            remove_page_breaks(paragraph)  # Удаляем разрыв страницы, но не весь параграф
        elif is_empty_paragraph(paragraph):
            remove_paragraph(paragraph)
            removed_after += 1
        else:
            break
    return removed_after
    


def addPageBreak(paragraph):
    """
    Добавляет разрыв страницы в конец предыдущего абзаца.
    Если перед параграфом уже есть разрыв страницы (либо как отдельный абзац,
    либо в конце предыдущего абзаца), новый не добавляется.
    """
    ns = {'w': 'http://schemas.openxmlformats.org/wordprocessingml/2006/main'}
    prev_paragraph = paragraph._element.getprevious()

    # Если предыдущего абзаца нет - выходим
    if prev_paragraph is None:
        return

    # Проверяем, содержит ли предыдущий абзац уже разрыв страницы
    # Ищем все элементы <w:br> с атрибутом w:type="page" в предыдущем абзаце
    page_breaks = prev_paragraph.findall(".//w:br[@w:type='page']", namespaces=ns)
    if page_breaks:
        # Если разрыв уже есть – ничего не делаем
        return

    # Если разрыва нет, добавляем его в конец предыдущего абзаца
    page_break = OxmlElement("w:br")
    page_break.set(qn("w:type"), "page")
    prev_paragraph.append(page_break)

def is_empty_paragraph_element(p_element):
    """
    Проверяет, является ли XML-элемент абзаца пустым.
    Пустым считается, если конкатенированный текст всех узлов пуст (после удаления пробелов).
    """
    text = ''.join(p_element.itertext()).strip()
    return text == ''

def addEmptyParagraphAfter(paragraph):
    """Добавляет пустой параграф сразу после указанного абзаца, если его там ещё нет."""
    if paragraph is None:
        return

    next_elem = paragraph._element.getnext()
    # Если существует следующий элемент и он является параграфом, проверяем его содержимое
    if next_elem is not None and next_elem.tag == qn("w:p") and is_empty_paragraph_element(next_elem):
        # Пустой абзац уже присутствует – ничего не делаем
        return

    # Иначе создаём новый пустой параграф и вставляем его сразу после текущего
    new_paragraph = OxmlElement("w:p")
    paragraph._element.addnext(new_paragraph)

def addEmptyParagraphBefore(paragraph):
    """Добавляет пустой параграф перед указанным абзацем, если его там ещё нет."""
    if paragraph is None:
        return

    prev_elem = paragraph._element.getprevious()
    # Если существует предыдущий элемент и он является параграфом, проверяем его содержимое
    if prev_elem is not None and prev_elem.tag == qn("w:p") and is_empty_paragraph_element(prev_elem):
        # Пустой абзац уже присутствует – ничего не делаем
        return

    # Иначе создаём новый пустой параграф и вставляем его перед текущим
    new_paragraph = OxmlElement("w:p")
    paragraph._element.addprevious(new_paragraph)

def ensureHeadingStyle(doc, level, font, fontsize):
    """Проверяет, создаёт или обновляет стиль заголовка."""
    base_style_name = f"Заголовок {level}"  # Русское название стиля
    
    # Проверяем существование стиля через XML
    for style in doc.styles:
        if style.type == WD_STYLE_TYPE.PARAGRAPH:
            style_elm = style._element
            name_elem = style_elm.find(qn("w:name"))
            if name_elem is not None and name_elem.get(qn("w:val")) and name_elem.get(qn("w:val")).startswith(base_style_name):
                return style.name  # Если стиль найден, возвращаем его имя
    
    # Создаём новый стиль, если не найден
    style_name = f"{base_style_name}{len([s for s in doc.styles if base_style_name in s.name]) + 1}"  # Добавляем номер к названию
    style = doc.styles.add_style(style_name, WD_STYLE_TYPE.PARAGRAPH)

    # Удаляем старые настройки шрифта
    for r in style.element.findall(qn("w:rPr")):
        style.element.remove(r)
    
    # Настраиваем шрифт и другие параметры
    style.font.name = font  # Шрифт
    style.font.size = Pt(float(fontsize))  # Размер шрифта
    style.font.bold = True  # Жирный текст
    style.font.color.rgb = RGBColor(0, 0, 0)  # Чёрный цвет
    
    # Настраиваем выравнивание
    style.paragraph_format.alignment = WD_PARAGRAPH_ALIGNMENT.LEFT
    
    # Добавляем уровень заголовка
    style_elm = style._element
    pPr = style_elm.find(qn("w:pPr"))
    if pPr is None:
        pPr = OxmlElement("w:pPr")
        style_elm.append(pPr)
    
    outline_lvl = pPr.find(qn("w:outlineLvl"))
    if outline_lvl is None:
        outline_lvl = OxmlElement("w:outlineLvl")
        pPr.append(outline_lvl)
    
    outline_lvl.set(qn("w:val"), str(level - 1))  # Word использует 0-основанные уровни
    
    # Делаем стиль видимым в списке стилей Word
    style.hidden = False
    style.quick_style = True  # Включает отображение в меню стилей
    
    return style_name  # Возвращаем имя стиля

def cycle_removeEmptyLinesAndPageBreaks(doc):
    """
    Вызывает функцию removeEmptyLinesAndPageBreaks(doc, index) до тех пор,
    пока документ не перестанет изменяться. Для проверки изменений сравниваются
    XML-представление документа и количество параграфов.
    """
    prev_xml = doc._element.xml
    prev_par_count = len(doc.paragraphs)
    total_removed_after = 0

    while True:
        for index, paragraph in enumerate(doc.paragraphs):
            isDraw = False
            level = headingLevel(paragraph.text)
            # Проверяем наличие элемента <w:drawing> или <w:pict>
            for run in paragraph.runs:
                drawing = run._element.find(qn("w:drawing"))
                pict = run._element.find(qn("w:pict"))
                if drawing is not None or pict is not None:
                    isDraw = True
                    break
            if not paragraph._element.xpath(".//w:numPr") and not isDraw and isHeading(paragraph) and level != False:
                removed_after = removeEmptyLinesAndPageBreaks(doc, index)
                total_removed_after += removed_after
                if level == 1:
                    addPageBreak(paragraph)
                    addEmptyParagraphAfter(paragraph)
                else:
                    if index > 0:
                        addEmptyParagraphBefore(paragraph)
                    addEmptyParagraphAfter(paragraph)
        current_xml = doc._element.xml
        current_par_count = len(doc.paragraphs)

        # Если документ не изменился ни по структуре, ни по количеству параграфов – завершаем цикл
        if current_xml == prev_xml and current_par_count == prev_par_count:
            break

        prev_xml = current_xml
        prev_par_count = current_par_count

    return total_removed_after

def changeNormalStyle(doc, font, fontsize, alignment, spacing, beforespacing, afterspacing, firstindentation):
    """
    Изменяет стиль Normal во всём документе, используя заданные параметры.
    """
    # Получаем стиль Normal
    style = doc.styles['Normal']

    # Изменяем настройки шрифта
    style.font.name = font
    style.font.size = Pt(float(fontsize))
    style.font.color.rgb = RGBColor(0, 0, 0)

    # Изменяем выравнивание абзаца
    if alignment == "По левому краю":
        style.paragraph_format.alignment = WD_PARAGRAPH_ALIGNMENT.LEFT
    elif alignment == "По центру":
        style.paragraph_format.alignment = WD_PARAGRAPH_ALIGNMENT.CENTER
    elif alignment == "По правому краю":
        style.paragraph_format.alignment = WD_PARAGRAPH_ALIGNMENT.RIGHT
    elif alignment == "По ширине":
        style.paragraph_format.alignment = WD_PARAGRAPH_ALIGNMENT.JUSTIFY

    # Междустрочный интервал
    style.paragraph_format.line_spacing = float(spacing)

    # Интервал перед абзацем
    style.paragraph_format.space_before = Pt(float(fontsize) * float(beforespacing))

    # Интервал после абзаца
    style.paragraph_format.space_after = Pt(float(fontsize) * float(afterspacing))

    # Сброс отступа всего абзаца к стандартному (без дополнительного левого отступа)
    style.paragraph_format.left_indent = Cm(0)

    # Отступ первой строки
    style.paragraph_format.first_line_indent = Cm(float(firstindentation))