import re
from docx.oxml import OxmlElement
from docx.oxml.ns import qn
from docx.shared import Pt, Cm, RGBColor
from docx.enum.text import WD_PARAGRAPH_ALIGNMENT
from docx.enum.style import WD_STYLE_TYPE

def headingLevel(text):
    """ Определяет уровень заголовка."""
    text_lower = text.lower()
    if (text_lower in ("содержание", "введение", "заключение", "реферат", "приложение")
        or (text_lower.startswith("список") and ("источников" in text_lower or "литературы" in text_lower))):
        return 1
    match = re.match(r"^\s*(\d+(?:\.\d+){0,3})(?=\s|$|[.)])", text)
    return min(match.group(1).count(".") + 1, 4) if match else False

def isHeading(paragraph):
    """ Проверяет, является ли абзац заголовком. """
    text = paragraph.text.strip()
    text_lower = text.lower()
    if not text or len(text) > 150:
        return False

    if (text_lower in ("содержание", "введение", "заключение", "реферат", "приложение")
        or (text_lower.startswith("список") and ("источников" in text_lower or "литературы" in text_lower))):
        return True
    if not paragraph.runs:
        return False

    bold_count = 0
    total_chars = 0
    paragraph_style = paragraph.style
    style_is_bold = False

    if paragraph_style and paragraph_style.font:
        style_is_bold = paragraph_style.font.bold
    for run in paragraph.runs:
        run_text = run.text.strip()
        if not run_text:
            continue

        total_chars += len(run_text)
        run_is_bold = run.bold if run.bold is not None else style_is_bold
        if run_is_bold:
            bold_count += len(run_text)
    if total_chars == 0:
        return False
    bold_percentage = (bold_count / total_chars) * 100
    if (bold_percentage >= 90 or total_chars - bold_count <= 3):
        return True
    return False


def addPageBreak(paragraph):
    """
    Добавляет разрыв страницы в конец предыдущего абзаца.
    Если перед параграфом уже есть разрыв страницы (либо как отдельный абзац,
    либо в конце предыдущего абзаца), новый не добавляется.
    """
    ns = {'w': 'http://schemas.openxmlformats.org/wordprocessingml/2006/main'}
    prev_paragraph = paragraph._element.getprevious()
    if prev_paragraph is None:
        return
    page_breaks = prev_paragraph.findall(".//w:br[@w:type='page']", namespaces=ns)
    if page_breaks:
        return
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
    if next_elem is not None and next_elem.tag == qn("w:p") and is_empty_paragraph_element(next_elem):
        return
    new_paragraph = OxmlElement("w:p")
    paragraph._element.addnext(new_paragraph)

def addEmptyParagraphBefore(paragraph):
    """Добавляет пустой параграф перед указанным абзацем, если его там ещё нет."""
    if paragraph is None:
        return

    prev_elem = paragraph._element.getprevious()
    if prev_elem is not None and prev_elem.tag == qn("w:p") and is_empty_paragraph_element(prev_elem):
        return
    new_paragraph = OxmlElement("w:p")
    paragraph._element.addprevious(new_paragraph)

def ensureHeadingStyle(doc, level, font, fontsize):
    """Проверяет, создаёт или обновляет стиль заголовка."""
    base_style_name = f"Заголовок {level}"
    for style in doc.styles:
        if style.type == WD_STYLE_TYPE.PARAGRAPH:
            style_elm = style._element
            name_elem = style_elm.find(qn("w:name"))
            if name_elem is not None and name_elem.get(qn("w:val")) and name_elem.get(qn("w:val")).startswith(base_style_name):
                return style.name
    style_name = f"{base_style_name}{len([s for s in doc.styles if base_style_name in s.name]) + 1}"
    style = doc.styles.add_style(style_name, WD_STYLE_TYPE.PARAGRAPH)
    for r in style.element.findall(qn("w:rPr")):
        style.element.remove(r)
    style.font.name = font
    style.font.size = Pt(float(fontsize))
    style.font.bold = True
    style.font.color.rgb = RGBColor(0, 0, 0)
    style.paragraph_format.alignment = WD_PARAGRAPH_ALIGNMENT.LEFT
    style_elm = style._element
    pPr = style_elm.find(qn("w:pPr"))
    if pPr is None:
        pPr = OxmlElement("w:pPr")
        style_elm.append(pPr)

    outline_lvl = pPr.find(qn("w:outlineLvl"))
    if outline_lvl is None:
        outline_lvl = OxmlElement("w:outlineLvl")
        pPr.append(outline_lvl)

    outline_lvl.set(qn("w:val"), str(level - 1))
    style.hidden = False
    style.quick_style = True

    return style_name

def changeNormalStyle(doc, font, fontsize, alignment, spacing, beforespacing, afterspacing, firstindentation):
    """
    Изменяет стиль Normal во всём документе, используя заданные параметры.
    """
    style = doc.styles['Normal']
    style.font.name = font
    style.font.size = Pt(float(fontsize))
    style.font.color.rgb = RGBColor(0, 0, 0)
    if alignment == "По левому краю":
        style.paragraph_format.alignment = WD_PARAGRAPH_ALIGNMENT.LEFT
    elif alignment == "По центру":
        style.paragraph_format.alignment = WD_PARAGRAPH_ALIGNMENT.CENTER
    elif alignment == "По правому краю":
        style.paragraph_format.alignment = WD_PARAGRAPH_ALIGNMENT.RIGHT
    elif alignment == "По ширине":
        style.paragraph_format.alignment = WD_PARAGRAPH_ALIGNMENT.JUSTIFY
    style.paragraph_format.line_spacing = float(spacing)
    style.paragraph_format.space_before = Pt(float(fontsize) * float(beforespacing))
    style.paragraph_format.space_after = Pt(float(fontsize) * float(afterspacing))
    style.paragraph_format.left_indent = Cm(0)
    style.paragraph_format.first_line_indent = Cm(float(firstindentation))
