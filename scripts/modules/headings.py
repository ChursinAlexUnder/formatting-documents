import re
from docx.oxml import OxmlElement
from docx.oxml.ns import qn
from collections import Counter
from docx.shared import Pt, RGBColor
from docx.enum.text import WD_PARAGRAPH_ALIGNMENT
from docx.enum.style import WD_STYLE_TYPE

def getDefaultFontSize(doc, fontsize):
    """Определяет самый часто встречающийся размер шрифта в документе по количеству символов."""
    font_sizes = Counter()

    for paragraph in doc.paragraphs:
        for run in paragraph.runs:
            text = run.text.strip()  # Игнорируем пустые строки (например, изображения)
            if text:  
                size = run.font.size
                if size is None and paragraph.style.font.size:
                    size = paragraph.style.font.size  # Используем стиль абзаца, если размер не задан
                
                if size:
                    font_sizes[size.pt] += len(text)  # Взвешиваем размер шрифта по количеству символов

    return font_sizes.most_common(1)[0][0] if font_sizes else fontsize

def isSurroundedEmptyLines(doc, index):
    """ Проверяет, окружён ли абзац полностью пустыми строками сверху и снизу (без текста и объектов). """
    
    def isTrulyEmpty(paragraph):
        """ Проверяет, является ли параграф полностью пустым (не содержит текста и объектов). """
        if paragraph.text.strip():  # Если есть текст, параграф не пустой
            return False
        
        # Проверяем, есть ли в runs хоть что-то, кроме текста (например, картинки)
        for run in paragraph.runs:
            if run.text.strip():  # Если в run есть текст, значит, параграф не пустой
                return False
            if run._element.getchildren():  # Если у run есть вложенные элементы (например, <w:drawing>), значит, там объект
                return False
        
        return True  # Если ничего нет, параграф считается пустым

    # Если это последний параграф, сразу возвращаем False
    if index == len(doc.paragraphs) - 1:
        return False
    
    # Проверяем, окружён ли параграф пустыми строками сверху и снизу
    return (index > 0 and isTrulyEmpty(doc.paragraphs[index - 1])) and \
           (index < len(doc.paragraphs) - 1 and isTrulyEmpty(doc.paragraphs[index + 1]))


def isBeforePageBreak(doc, index):
    """Проверяет, есть ли перед данным абзацем разрыв страницы.
       Если сам абзац пустой, возвращает False.
       Но если между абзацем и разрывом есть пустые строки, то у этого абзаца будет True.
    """
    ns = {'w': 'http://schemas.openxmlformats.org/wordprocessingml/2006/main'}

    # Если сам абзац пустой, сразу возвращаем False.
    if not doc.paragraphs[index].text.strip():
        return False

    # Идем назад по абзацам от текущего до начала документа.
    for i in range(index - 1, -1, -1):
        paragraph = doc.paragraphs[i]
        # Проходим по каждому run в абзаце
        for run in paragraph.runs:
            # Ищем элемент <w:br> с атрибутом w:type="page"
            brs = run._element.findall(".//w:br", namespaces=ns)
            for br in brs:
                if br.get(qn("w:type")) == "page":
                    return True
        # Если встретили абзац с каким-либо текстом, считаем, что разрыва страницы перед нужным абзацем нет.
        if paragraph.text.strip():
            return False
    return False

def headingLevel(text):
    """ Определяет уровень заголовка."""
    match = re.match(r"^\d+(\.\d+){0,3}", text)
    return min(len(match.group().split(".")), 4) if match else 1

def isHeading(paragraph, index, doc, defaultFontSize):
    """ Проверяет, является ли абзац заголовком. """
    text = paragraph.text.strip()
    
    # Проверка на пустой текст или слишком длинный абзац
    if not text or len(text) > 150:
        return False

    # Проверка, если абзац содержит только картинку/формулу
    if not paragraph.runs:  # Если нет run'ов, значит, текст отсутствует (например, изображение)
        return False

    bold_count = 0
    large_font_count = 0
    total_chars = 0

    # Получаем стиль параграфа (если есть)
    paragraph_style = paragraph.style
    style_is_bold = False
    style_font_size = None

    if paragraph_style and paragraph_style.font:
        style_is_bold = paragraph_style.font.bold
        style_font_size = paragraph_style.font.size.pt if paragraph_style.font.size else None

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

        # Определяем размер шрифта с учётом стиля параграфа
        run_font_size = run.font.size.pt if run.font.size else style_font_size
        if run_font_size and run_font_size > defaultFontSize:
            large_font_count += len(run_text)

    # Если нет текста (например, все run'ы пустые), то это не заголовок
    if total_chars == 0:
        return False

    # Проверяем процентное соотношение
    bold_percentage = (bold_count / total_chars) * 100
    large_font_percentage = (large_font_count / total_chars) * 100

    # Условия определения заголовка
    if (bold_percentage >= 90 or 
        large_font_percentage >= 90 or 
        total_chars - bold_count <= 3 or 
        total_chars - large_font_count <= 3):
        return True
    return isSurroundedEmptyLines(doc, index) or isBeforePageBreak(doc, index)

def removeEmptyLinesAndPageBreaks(doc, index):
    """Удаляет пустые строки, пустые абзацы и разрывы страниц перед и после параграфа, не затрагивая контент (картинки, таблицы)."""
    ns = {'w': 'http://schemas.openxmlformats.org/wordprocessingml/2006/main'}

    def is_empty_paragraph(paragraph):
        """
        Проверяет, является ли абзац пустым.
        Абзац считается непустым, если:
          - содержит не только пробельные символы, или
          - содержит встроенные объекты (например, картинки).
        """
        # Если есть текст, отличающийся от пробелов, считаем абзац не пустым.
        if paragraph.text and paragraph.text.strip():
            return False

        # Если в абзаце присутствуют встроенные объекты (например, картинки),
        # то он не должен удаляться.
        if paragraph._element.find('.//w:drawing', namespaces=ns) is not None:
            return False

        # Если абзац содержит только разрывы страниц, пробелы или вообще ничего – он пустой.
        return True

    def remove_paragraph(paragraph):
        """Корректно удаляет абзац из документа."""
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

    # Удаление пустых абзацев и разрывов страниц после указанного индекса
    i = index + 1
    while i < len(doc.paragraphs):
        paragraph = doc.paragraphs[i]
        page_break = paragraph._element.find(".//w:br[@w:type='page']", namespaces=ns)

        if page_break is not None:
            remove_page_breaks(paragraph)  # Удаляем разрыв страницы, но не весь параграф
        elif is_empty_paragraph(paragraph):
            remove_paragraph(paragraph)
        else:
            break

def addPageBreak(paragraph):
    """Добавляет разрыв страницы в конец предыдущего абзаца.  
    Если параграф первый в документе, создаёт разрыв страницы отдельным абзацем."""
    
    prev_paragraph = paragraph._element.getprevious()

    if prev_paragraph is not None:
        # Если перед параграфом уже есть другой абзац, добавляем разрыв страницы в его конец
        page_break = OxmlElement("w:br")
        page_break.set(qn("w:type"), "page")
        prev_paragraph.append(page_break)

def addEmptyParagraphAfter(paragraph):
    """Добавляет пустой параграф сразу после указанного абзаца."""
    if paragraph is None:
        return
    
    # Создаём новый пустой параграф (XML-элемент <w:p>)
    new_paragraph = OxmlElement("w:p")
    # Вставляем его сразу после переданного абзаца
    paragraph._element.addnext(new_paragraph)

def addEmptyParagraphBefore(paragraph):
    """Добавляет пустой параграф перед указанным абзацем."""
    if paragraph is None:
        return
    
    # Создаём новый пустой параграф (XML-элемент <w:p>)
    new_paragraph = OxmlElement("w:p")
    # Вставляем его перед переданным абзацем
    paragraph._element.addprevious(new_paragraph)

def ensureHeadingStyle(doc, level, font, fontsize):
    """Проверяет, создаёт или обновляет стиль заголовка."""
    style_name = f"Heading {level}"  # Название стиля, как в Word

    # Проверяем, есть ли уже такой стиль
    if style_name in doc.styles:
        style = doc.styles[style_name]  # Если есть, получаем стиль
    else:
        style = doc.styles.add_style(style_name, WD_STYLE_TYPE.PARAGRAPH)  # Если нет, создаём

    # Получаем объект run, чтобы корректно менять шрифт (исправляет баг с невидимым шрифтом)
    for r in style.element.findall(qn("w:rPr")):
        style.element.remove(r)  # Удаляем старые настройки шрифта (чтобы обновились)

    # Настраиваем шрифт и другие параметры (вне зависимости от существования стиля)
    style.font.name = font  # Шрифт
    style.font.size = Pt(float(fontsize))  # Размер шрифта
    style.font.bold = True  # Жирный текст
    style.font.color.rgb = RGBColor(0, 0, 0)  # Чёрный цвет

    # Настраиваем выравнивание
    style.paragraph_format.alignment = WD_PARAGRAPH_ALIGNMENT.LEFT

    # Добавляем уровень заголовка (чтобы отображался в оглавлении и сворачивался)
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